package repository

import (
	"context"

	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReactionRepository defines the interface for reaction data operations.
type ReactionRepository interface {
	AddReaction(ctx context.Context, reaction *domain.Reaction) error
	RemoveReaction(ctx context.Context, userID, postID primitive.ObjectID, reactionType domain.ReactionType) error
	HasReacted(ctx context.Context, userID, postID primitive.ObjectID, reactionType domain.ReactionType) (bool, error)
}

type mongoReactionRepository struct {
	db *mongo.Database
}

// NewMongoReactionRepository creates a new reaction repository.
func NewMongoReactionRepository(db *mongo.Database) ReactionRepository {
	return &mongoReactionRepository{db: db}
}

func (r *mongoReactionRepository) reactions() *mongo.Collection {
	return r.db.Collection("reactions")
}
func (r *mongoReactionRepository) posts() *mongo.Collection {
	return r.db.Collection("posts")
}

func (r *mongoReactionRepository) AddReaction(ctx context.Context, reaction *domain.Reaction) error {
	// Transactional logic
	filter := bson.M{"userid": reaction.UserID, "postid": reaction.PostID, "type": reaction.Type}
	update := bson.M{"$setOnInsert": reaction}
	opts := options.Update().SetUpsert(true)
	res, err := r.reactions().UpdateOne(ctx, filter, update, opts)
	if err != nil || res.UpsertedCount == 0 {
		return err // Error or reaction already exists
	}

	// Increment the correct counter on the post
	_, err = r.posts().UpdateOne(ctx, bson.M{"_id": reaction.PostID}, bson.M{"$inc": bson.M{string(reaction.Type) + "count": 1}})
	return err
}

func (r *mongoReactionRepository) RemoveReaction(ctx context.Context, userID, postID primitive.ObjectID, reactionType domain.ReactionType) error {
	// Transactional logic
	res, err := r.reactions().DeleteOne(ctx, bson.M{"userid": userID, "postid": postID, "type": reactionType})
	if err != nil || res.DeletedCount == 0 {
		return err // Error or reaction didn't exist
	}

	// Decrement the correct counter on the post
	_, err = r.posts().UpdateOne(ctx, bson.M{"_id": postID}, bson.M{"$inc": bson.M{string(reactionType) + "count": -1}})
	return err
}

func (r *mongoReactionRepository) HasReacted(ctx context.Context, userID, postID primitive.ObjectID, reactionType domain.ReactionType) (bool, error) {
	count, err := r.reactions().CountDocuments(ctx, bson.M{"userid": userID, "postid": postID, "type": reactionType})
	return count > 0, err
}
