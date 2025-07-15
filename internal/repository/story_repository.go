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
// Stories are temporary content that expire after 24 hours and are
// only visible to users who follow the story creator.
type StoryRepository interface {
	// CreateStory creates a new story in the database
	CreateStory(ctx context.Context, story *domain.Story) error
	// GetStoriesByUserID retrieves all active stories for a specific user
	GetStoriesByUserID(ctx context.Context, userID primitive.ObjectID) ([]domain.Story, error)
	// GetStoriesForFeed retrieves stories from followed users for the feed
	GetStoriesForFeed(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.Story, error)
	// DeleteStory removes a story from the database
	DeleteStory(ctx context.Context, storyID, userID primitive.ObjectID) error
	// DeleteExpiredStories removes all stories that have expired (older than 24 hours)
	DeleteExpiredStories(ctx context.Context) error
	FindExpired(ctx context.Context) ([]domain.Story, error)
	DeleteMany(ctx context.Context, storyIDs []primitive.ObjectID) error
}

// mongoStoryRepository implements StoryRepository using MongoDB as the backend
type mongoStoryRepository struct {
	collection *mongo.Collection
}

// NewMongoStoryRepository creates a new story repository instance with MongoDB backend.
// The repository handles all story-related database operations including CRUD operations
// and automatic cleanup of expired stories.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - StoryRepository: A configured story repository ready for use
func NewMongoStoryRepository(db *mongo.Database) StoryRepository {
	return &mongoStoryRepository{
		collection: db.Collection("stories"),
	}
}

func (r *mongoStoryRepository) CreateStory(ctx context.Context, story *domain.Story) error {
	_, err := r.collection.InsertOne(ctx, story)
	return err
}

func (r *mongoStoryRepository) GetStoriesByUserID(ctx context.Context, userID primitive.ObjectID) ([]domain.Story, error) {
	var stories []domain.Story
	cursor, err := r.collection.Find(ctx, bson.M{"userid": userID, "expiresat": bson.M{"$gt": time.Now()}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &stories)
	return stories, err
}

func (r *mongoStoryRepository) GetStoriesForFeed(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.Story, error) {
	var stories []domain.Story
	cursor, err := r.collection.Find(ctx, bson.M{"userid": bson.M{"$in": userIDs}, "expiresat": bson.M{"$gt": time.Now()}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &stories)
	return stories, err
}

func (r *mongoStoryRepository) DeleteStory(ctx context.Context, storyID, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": storyID, "userid": userID})
	return err
}

func (r *mongoStoryRepository) DeleteExpiredStories(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"expiresat": bson.M{"$lte": time.Now()}})
	return err
}

func (r *mongoStoryRepository) FindExpired(ctx context.Context) ([]domain.Story, error) {
	var stories []domain.Story
	cursor, err := r.collection.Find(ctx, bson.M{"expiresat": bson.M{"$lte": time.Now()}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &stories)
	return stories, err
}

func (r *mongoStoryRepository) DeleteMany(ctx context.Context, storyIDs []primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": storyIDs}})
	return err
}
