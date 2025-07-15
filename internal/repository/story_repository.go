package repository

import (
	"context"
	"time"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	GetStoriesForFeed(ctx context.Context, userID primitive.ObjectID) ([]domain.Story, error)
	// DeleteStory removes a story from the database
	DeleteStory(ctx context.Context, storyID, userID primitive.ObjectID) error
	// DeleteExpiredStories removes all stories that have expired (older than 24 hours)
	DeleteExpiredStories(ctx context.Context) error
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
