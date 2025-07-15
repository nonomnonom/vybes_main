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
	GetFollowingIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
	GetFollowerIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
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

func (r *mongoFollowRepository) CreateFollow(ctx context.Context, follow *domain.Follow) error {
	_, err := r.collection.InsertOne(ctx, follow)
	return err
}

func (r *mongoFollowRepository) DeleteFollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"followerid": followerID, "followingid": followingID})
	return err
}

func (r *mongoFollowRepository) GetFollowers(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Follow, error) {
	var follows []domain.Follow
	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"followingid": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &follows)
	return follows, err
}

func (r *mongoFollowRepository) GetFollowing(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Follow, error) {
	var follows []domain.Follow
	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"followerid": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &follows)
	return follows, err
}

func (r *mongoFollowRepository) GetFollowingIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var follows []domain.Follow
	cursor, err := r.collection.Find(ctx, bson.M{"followerid": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, len(follows))
	for i, f := range follows {
		ids[i] = f.FollowingID
	}
	return ids, nil
}

func (r *mongoFollowRepository) GetFollowerIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var follows []domain.Follow
	cursor, err := r.collection.Find(ctx, bson.M{"followingid": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, len(follows))
	for i, f := range follows {
		ids[i] = f.FollowerID
	}
	return ids, nil
}

func (r *mongoFollowRepository) IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"followerid": followerID, "followingid": followingID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoFollowRepository) GetFollowerCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"followingid": userID})
}

func (r *mongoFollowRepository) GetFollowingCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"followerid": userID})
}
