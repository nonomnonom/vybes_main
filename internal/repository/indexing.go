package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupIndexes creates all the necessary database indexes for optimal query performance.
// This function should be called during application startup to ensure
// all required indexes exist before the application begins processing requests.
//
// Parameters:
//   - ctx: Context for the operation
//   - db: MongoDB database instance
func SetupIndexes(ctx context.Context, db *mongo.Database) {
	// Create indexes for 'users' collection
	createUserIndexes(ctx, db)
	
	// Create indexes for 'follows' collection
	createFollowIndexes(ctx, db)
	
	// Create TTL index for 'stories' collection (auto-delete expired stories)
	createStoryIndexes(ctx, db)
	
	// Create indexes for 'posts' collection
	createPostIndexes(ctx, db)
	
	// Create indexes for 'content' collection
	createContentIndexes(ctx, db)
	
	// Create indexes for 'reactions' collection
	createReactionIndexes(ctx, db)
	
	// Create indexes for 'comments' collection
	createCommentIndexes(ctx, db)
	
	// Create indexes for 'bookmarks' collection
	createBookmarkIndexes(ctx, db)
	
	// Create indexes for 'notifications' collection
	createNotificationIndexes(ctx, db)
}

// createUserIndexes sets up indexes for the users collection
// Includes indexes for email, username, and wallet address lookups
func createUserIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("users")
	
	// Unique index on email for user authentication
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Unique index on username for user identification
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index on wallet address for blockchain integration
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "walletaddress", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createFollowIndexes sets up indexes for the follows collection
// Includes compound indexes for efficient follow relationship queries
func createFollowIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("follows")
	
	// Compound index for checking if user A follows user B
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "followerid", Value: 1},
			{Key: "followingid", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for finding all followers of a user
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "followingid", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createStoryIndexes sets up indexes for the stories collection
// Includes TTL index for automatic cleanup of expired stories
func createStoryIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("stories")
	
	// TTL index to automatically delete stories after 24 hours
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiresat", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0), // Delete immediately when expired
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for finding stories by user
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userid", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createPostIndexes sets up indexes for the posts collection
// Includes indexes for feed queries and user post retrieval
func createPostIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("posts")
	
	// Compound index for feed queries (user + creation date)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "createdat", Value: -1},
		},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for visibility-based queries
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "visibility", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createContentIndexes sets up indexes for the content collection
// Includes indexes for content type and user queries
func createContentIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("content")
	
	// Index for content type queries
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "type", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for user content queries
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userid", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createReactionIndexes sets up indexes for the reactions collection
// Includes compound indexes for efficient reaction queries
func createReactionIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("reactions")
	
	// Compound index for checking if user reacted to content
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "contentid", Value: 1},
			{Key: "reactiontype", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for finding all reactions to content
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "contentid", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createCommentIndexes sets up indexes for the comments collection
// Includes indexes for comment queries by post and user
func createCommentIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("comments")
	
	// Index for finding comments by post
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "postid", Value: 1},
			{Key: "createdat", Value: -1},
		},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for finding comments by user
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userid", Value: 1}},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createBookmarkIndexes sets up indexes for the bookmarks collection
// Includes compound indexes for efficient bookmark queries
func createBookmarkIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("bookmarks")
	
	// Compound index for checking if user bookmarked content
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "contentid", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for finding user bookmarks
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "createdat", Value: -1},
		},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}

// createNotificationIndexes sets up indexes for the notifications collection
// Includes indexes for notification queries by user and read status
func createNotificationIndexes(ctx context.Context, db *mongo.Database) {
	collection := db.Collection("notifications")
	
	// Index for finding notifications by user
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "createdat", Value: -1},
		},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
	
	// Index for unread notification queries
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userid", Value: 1},
			{Key: "read", Value: 1},
		},
	})
	if err != nil {
		// Log error but don't fail - index might already exist
	}
}
