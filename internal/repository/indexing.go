package repository

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupIndexes creates all the necessary indexes for the collections.
func SetupIndexes(ctx context.Context, db *mongo.Database) {
	log.Info().Msg("Setting up database indexes...")

	// Indexes for 'users' collection
	usersCollection := db.Collection("users")
	userIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "vid", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
		{Keys: bson.D{{Key: "name", Value: "text"}, {Key: "username", Value: "text"}}},
	}
	createIndexes(ctx, usersCollection, userIndexes, "users")

	// Indexes for 'follows' collection
	followsCollection := db.Collection("follows")
	followIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "followerid", Value: 1}, {Key: "followingid", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "followerid", Value: 1}}},
		{Keys: bson.D{{Key: "followingid", Value: 1}}},
	}
	createIndexes(ctx, followsCollection, followIndexes, "follows")

	// TTL Index for 'stories' collection
	storiesCollection := db.Collection("stories")
	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresat", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	createIndexes(ctx, storiesCollection, []mongo.IndexModel{ttlIndex}, "stories")

	// Indexes for 'posts' collection
	postsCollection := db.Collection("posts")
	postIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userid", Value: 1}}},
		{Keys: bson.D{{Key: "contentid", Value: 1}}},
		{Keys: bson.D{{Key: "createdat", Value: -1}}},
	}
	createIndexes(ctx, postsCollection, postIndexes, "posts")

	// Indexes for 'content' collection
	contentCollection := db.Collection("content")
	contentIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userid", Value: 1}}},
	}
	createIndexes(ctx, contentCollection, contentIndexes, "content")

	// Indexes for 'reactions' collection
	reactionsCollection := db.Collection("reactions")
	reactionIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userid", Value: 1}, {Key: "postid", Value: 1}, {Key: "type", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "postid", Value: 1}}},
	}
	createIndexes(ctx, reactionsCollection, reactionIndexes, "reactions")

	// Indexes for 'comments' collection
	commentsCollection := db.Collection("comments")
	commentIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "postid", Value: 1}}},
		{Keys: bson.D{{Key: "createdat", Value: -1}}},
	}
	createIndexes(ctx, commentsCollection, commentIndexes, "comments")

	// Indexes for 'bookmarks' collection
	bookmarksCollection := db.Collection("bookmarks")
	bookmarkIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userid", Value: 1}, {Key: "postid", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "userid", Value: 1}, {Key: "createdat", Value: -1}}},
	}
	createIndexes(ctx, bookmarksCollection, bookmarkIndexes, "bookmarks")

	// Indexes for 'notifications' collection
	notificationsCollection := db.Collection("notifications")
	notificationIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userid", Value: 1}, {Key: "createdat", Value: -1}}},
	}
	createIndexes(ctx, notificationsCollection, notificationIndexes, "notifications")

	log.Info().Msg("Database indexes setup complete.")
}

func createIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel, collectionName string) {
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		log.Warn().Err(err).Str("collection", collectionName).Msg("Could not create indexes (they may already exist)")
	} else {
		log.Info().Str("collection", collectionName).Msg("Indexes created successfully")
	}
}