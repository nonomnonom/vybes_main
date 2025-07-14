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
type ContentRepository interface {
	CreateContent(ctx context.Context, content *domain.Content) error
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, postID primitive.ObjectID) (*domain.Post, error)
	GetPostsByIDs(ctx context.Context, postIDs []primitive.ObjectID) ([]domain.Post, error)
	GetPostsByUsersWithVisibility(ctx context.Context, userIDs []primitive.ObjectID, visibilities []domain.PostVisibility, limit int) ([]domain.Post, error)
	GetRepostsByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Post, error)
	CreateComment(ctx context.Context, comment *domain.Comment) error
	GetCommentsByPost(ctx context.Context, postID primitive.ObjectID, limit int) ([]domain.Comment, error)
	IncrementCommentCount(ctx context.Context, postID primitive.ObjectID) error
	IncrementRepostCount(ctx context.Context, postID primitive.ObjectID) error
	IncrementViewCount(ctx context.Context, postID primitive.ObjectID) error
}

type mongoContentRepository struct {
	db *mongo.Database
}

// NewMongoContentRepository creates a new content repository.
func NewMongoContentRepository(db *mongo.Database) ContentRepository {
	return &mongoContentRepository{db: db}
}

// Collection helpers
func (r *mongoContentRepository) posts() *mongo.Collection {
	return r.db.Collection("posts")
}
func (r *mongoContentRepository) content() *mongo.Collection {
	return r.db.Collection("content")
}
func (r *mongoContentRepository) comments() *mongo.Collection {
	return r.db.Collection("comments")
}

// Content and Post methods
func (r *mongoContentRepository) CreateContent(ctx context.Context, content *domain.Content) error {
	_, err := r.content().InsertOne(ctx, content)
	return err
}

func (r *mongoContentRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	_, err := r.posts().InsertOne(ctx, post)
	return err
}

func (r *mongoContentRepository) GetPostByID(ctx context.Context, postID primitive.ObjectID) (*domain.Post, error) {
	var post domain.Post
	err := r.posts().FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *mongoContentRepository) GetPostsByIDs(ctx context.Context, postIDs []primitive.ObjectID) ([]domain.Post, error) {
	filter := bson.M{"_id": bson.M{"$in": postIDs}}
	cursor, err := r.posts().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []domain.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *mongoContentRepository) GetPostsByUsersWithVisibility(ctx context.Context, userIDs []primitive.ObjectID, visibilities []domain.PostVisibility, limit int) ([]domain.Post, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit))
	filter := bson.M{
		"userid":     bson.M{"$in": userIDs},
		"visibility": bson.M{"$in": visibilities},
	}
	cursor, err := r.posts().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []domain.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *mongoContentRepository) GetRepostsByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]domain.Post, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit))
	filter := bson.M{
		"userid":         userID,
		"originalpostid": bson.M{"$ne": nil},
	}
	cursor, err := r.posts().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []domain.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *mongoContentRepository) IncrementRepostCount(ctx context.Context, postID primitive.ObjectID) error {
	_, err := r.posts().UpdateOne(ctx, bson.M{"_id": postID}, bson.M{"$inc": bson.M{"repostcount": 1}})
	return err
}

func (r *mongoContentRepository) IncrementViewCount(ctx context.Context, postID primitive.ObjectID) error {
	_, err := r.posts().UpdateOne(ctx, bson.M{"_id": postID}, bson.M{"$inc": bson.M{"viewcount": 1}})
	return err
}

// Comment methods
func (r *mongoContentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	_, err := r.comments().InsertOne(ctx, comment)
	return err
}

func (r *mongoContentRepository) IncrementCommentCount(ctx context.Context, postID primitive.ObjectID) error {
	_, err := r.posts().UpdateOne(ctx, bson.M{"_id": postID}, bson.M{"$inc": bson.M{"commentcount": 1}})
	return err
}

func (r *mongoContentRepository) GetCommentsByPost(ctx context.Context, postID primitive.ObjectID, limit int) ([]domain.Comment, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit))
	filter := bson.M{"postid": postID}
	cursor, err := r.comments().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []domain.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}