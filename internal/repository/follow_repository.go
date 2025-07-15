package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FollowRepository defines the interface for follow relationship data operations.
// Follow relationships represent user connections where one user follows another,
// enabling content visibility and social interactions.
type FollowRepository interface {
	// CreateFollow establishes a follow relationship between two users
	CreateFollow(ctx context.Context, follow *domain.Follow) error
	// DeleteFollow removes a follow relationship between two users
	DeleteFollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	// GetFollowers retrieves all users who follow a specific user
	GetFollowers(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Follow, error)
	// GetFollowing retrieves all users that a specific user is following
	GetFollowing(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Follow, error)
	// IsFollowing checks if one user is following another
	IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error)
	// GetFollowerCount returns the number of followers for a user
	GetFollowerCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	// GetFollowingCount returns the number of users a user is following
	GetFollowingCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
}

// mongoFollowRepository implements FollowRepository using MongoDB as the backend
type mongoFollowRepository struct {
	collection *mongo.Collection
}

// NewMongoFollowRepository creates a new follow repository instance with MongoDB backend.
// The repository handles all follow relationship database operations including
// creating, deleting, and querying user follow relationships.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - FollowRepository: A configured follow repository ready for use
func NewMongoFollowRepository(db *mongo.Database) FollowRepository {
	return &mongoFollowRepository{
		collection: db.Collection("follows"),
	}
}
