package repository

import (
	"context"
	"time"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StoryRepository defines the interface for story data operations.
type StoryRepository interface {
	Create(ctx context.Context, story *domain.Story) error
	GetStoriesByUsers(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.Story, error)
	FindExpired(ctx context.Context) ([]domain.Story, error)
	DeleteMany(ctx context.Context, storyIDs []primitive.ObjectID) error
}

type mongoStoryRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoStoryRepository creates a new story repository with MongoDB.
func NewMongoStoryRepository(db *mongo.Database) StoryRepository {
	return &mongoStoryRepository{
		db:         db,
		collection: "stories",
	}
}

func (r *mongoStoryRepository) Create(ctx context.Context, story *domain.Story) error {
	_, err := r.db.Collection(r.collection).InsertOne(ctx, story)
	return err
}

func (r *mongoStoryRepository) GetStoriesByUsers(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.Story, error) {
	filter := bson.M{"userid": bson.M{"$in": userIDs}}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []domain.Story
	if err = cursor.All(ctx, &stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *mongoStoryRepository) FindExpired(ctx context.Context) ([]domain.Story, error) {
	filter := bson.M{"expiresat": bson.M{"$lte": time.Now()}}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []domain.Story
	if err = cursor.All(ctx, &stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *mongoStoryRepository) DeleteMany(ctx context.Context, storyIDs []primitive.ObjectID) error {
	filter := bson.M{"_id": bson.M{"$in": storyIDs}}
	_, err := r.db.Collection(r.collection).DeleteMany(ctx, filter)
	return err
}