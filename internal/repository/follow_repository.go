package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vybes/internal/domain"
)

// FollowRepository defines the interface for follow data operations.
type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error)
	GetFollowerCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	GetFollowingCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	GetFollowingIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
	GetFollowerIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type mongoFollowRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoFollowRepository creates a new follow repository with MongoDB.
func NewMongoFollowRepository(db *mongo.Database) FollowRepository {
	return &mongoFollowRepository{
		db:         db,
		collection: "follows",
	}
}

func (r *mongoFollowRepository) Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	filter := bson.M{"followerid": followerID, "followingid": followingID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":         primitive.NewObjectID(),
			"followerid":  followerID,
			"followingid": followingID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *mongoFollowRepository) Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	filter := bson.M{"followerid": followerID, "followingid": followingID}
	_, err := r.db.Collection(r.collection).DeleteOne(ctx, filter)
	return err
}

func (r *mongoFollowRepository) IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error) {
	filter := bson.M{"followerid": followerID, "followingid": followingID}
	count, err := r.db.Collection(r.collection).CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoFollowRepository) GetFollowerCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{"followingid": userID}
	return r.db.Collection(r.collection).CountDocuments(ctx, filter)
}

func (r *mongoFollowRepository) GetFollowingCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{"followerid": userID}
	return r.db.Collection(r.collection).CountDocuments(ctx, filter)
}

func (r *mongoFollowRepository) GetFollowingIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.M{"followerid": userID}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var followingIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var follow domain.Follow
		if err := cursor.Decode(&follow); err == nil {
			followingIDs = append(followingIDs, follow.FollowingID)
		}
	}
	return followingIDs, nil
}

func (r *mongoFollowRepository) GetFollowerIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.M{"followingid": userID}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var followerIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var follow domain.Follow
		if err := cursor.Decode(&follow); err == nil {
			followerIDs = append(followerIDs, follow.FollowerID)
		}
	}
	return followerIDs, nil
}