package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ContentRepository defines the interface for content and post data operations.
// Content represents the core data entities in the application including
// posts, comments, and associated metadata.
type ContentRepository interface {
	// Content and Post methods
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, postID primitive.ObjectID) (*domain.Post, error)
	GetPostsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Post, error)
	GetPostsByIDs(ctx context.Context, postIDs []primitive.ObjectID) ([]domain.Post, error)
	GetPostsByUsersWithVisibility(ctx context.Context, userIDs []primitive.ObjectID, visibilities []domain.PostVisibility, limit int) ([]domain.Post, error)
	UpdatePost(ctx context.Context, post *domain.Post) error
	DeletePost(ctx context.Context, postID, userID primitive.ObjectID) error
	GetFeedPosts(ctx context.Context, userIDs []primitive.ObjectID, page, limit int) ([]domain.Post, error)
	
	// Comment methods
	CreateComment(ctx context.Context, comment *domain.Comment) error
	GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	GetCommentCount(ctx context.Context, postID primitive.ObjectID) (int64, error)
}

// mongoContentRepository implements ContentRepository using MongoDB as the backend
type mongoContentRepository struct {
	postsCollection    *mongo.Collection
	commentsCollection *mongo.Collection
}

// NewMongoContentRepository creates a new content repository instance with MongoDB backend.
// The repository handles all content-related database operations including
// posts, comments, and their associated metadata.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - ContentRepository: A configured content repository ready for use
func NewMongoContentRepository(db *mongo.Database) ContentRepository {
	return &mongoContentRepository{
		postsCollection:    db.Collection("posts"),
		commentsCollection: db.Collection("comments"),
	}
}

// Collection helpers provide access to the underlying MongoDB collections
func (r *mongoContentRepository) posts() *mongo.Collection {
	return r.postsCollection
}

func (r *mongoContentRepository) comments() *mongo.Collection {
	return r.commentsCollection
}

func (r *mongoContentRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	_, err := r.posts().InsertOne(ctx, post)
	return err
}

func (r *mongoContentRepository) GetPostByID(ctx context.Context, postID primitive.ObjectID) (*domain.Post, error) {
	var post domain.Post
	err := r.posts().FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	return &post, err
}

func (r *mongoContentRepository) GetPostsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Post, error) {
	var posts []domain.Post
	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.posts().Find(ctx, bson.M{"userid": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &posts)
	return posts, err
}

func (r *mongoContentRepository) GetPostsByIDs(ctx context.Context, postIDs []primitive.ObjectID) ([]domain.Post, error) {
	var posts []domain.Post
	cursor, err := r.posts().Find(ctx, bson.M{"_id": bson.M{"$in": postIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &posts)
	return posts, err
}

func (r *mongoContentRepository) GetPostsByUsersWithVisibility(ctx context.Context, userIDs []primitive.ObjectID, visibilities []domain.PostVisibility, limit int) ([]domain.Post, error) {
	var posts []domain.Post
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdat", Value: -1}})
	cursor, err := r.posts().Find(ctx, bson.M{"userid": bson.M{"$in": userIDs}, "visibility": bson.M{"$in": visibilities}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &posts)
	return posts, err
}

func (r *mongoContentRepository) UpdatePost(ctx context.Context, post *domain.Post) error {
	_, err := r.posts().UpdateOne(ctx, bson.M{"_id": post.ID}, bson.M{"$set": post})
	return err
}

func (r *mongoContentRepository) DeletePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	_, err := r.posts().DeleteOne(ctx, bson.M{"_id": postID, "userid": userID})
	return err
}

func (r *mongoContentRepository) GetFeedPosts(ctx context.Context, userIDs []primitive.ObjectID, page, limit int) ([]domain.Post, error) {
	var posts []domain.Post
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.posts().Find(ctx, bson.M{"userid": bson.M{"$in": userIDs}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &posts)
	return posts, err
}

func (r *mongoContentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	_, err := r.comments().InsertOne(ctx, comment)
	return err
}

func (r *mongoContentRepository) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int) ([]domain.Comment, error) {
	var comments []domain.Comment
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.comments().Find(ctx, bson.M{"postid": postID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &comments)
	return comments, err
}

func (r *mongoContentRepository) DeleteComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	_, err := r.comments().DeleteOne(ctx, bson.M{"_id": commentID, "userid": userID})
	return err
}

func (r *mongoContentRepository) GetCommentCount(ctx context.Context, postID primitive.ObjectID) (int64, error) {
	return r.comments().CountDocuments(ctx, bson.M{"postid": postID})
}
