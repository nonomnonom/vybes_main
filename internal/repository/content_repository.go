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
