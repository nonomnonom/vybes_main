package repository

import (
	"context"
	"fmt"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReactionRepository defines the interface for reaction data operations.
// Reactions represent user interactions with content such as likes, loves,
// and other emotional responses to posts and comments.
type ReactionRepository interface {
	// AddReaction creates a new reaction for a specific content item
	AddReaction(ctx context.Context, reaction *domain.Reaction) error
	// RemoveReaction deletes an existing reaction from a content item
	RemoveReaction(ctx context.Context, userID, contentID primitive.ObjectID, reactionType domain.ReactionType) error
	// GetReactionsByContentID retrieves all reactions for a specific content item
	GetReactionsByContentID(ctx context.Context, contentID primitive.ObjectID) ([]domain.Reaction, error)
	// GetReactionCounts returns the count of each reaction type for a content item
	GetReactionCounts(ctx context.Context, contentID primitive.ObjectID) (map[domain.ReactionType]int64, error)
	// HasUserReacted checks if a user has reacted to a specific content item
	HasUserReacted(ctx context.Context, userID, contentID primitive.ObjectID, reactionType domain.ReactionType) (bool, error)
}

// mongoReactionRepository implements ReactionRepository using MongoDB as the backend
type mongoReactionRepository struct {
	collection *mongo.Collection
}

// NewMongoReactionRepository creates a new reaction repository instance with MongoDB backend.
// The repository handles all reaction-related database operations including adding,
// removing, and querying user reactions to content.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - ReactionRepository: A configured reaction repository ready for use
func NewMongoReactionRepository(db *mongo.Database) ReactionRepository {
	return &mongoReactionRepository{
		collection: db.Collection("reactions"),
	}
}

// AddReaction creates a new reaction in the database and updates the content's reaction counters.
// This operation is performed within a transaction to ensure data consistency.
//
// Parameters:
//   - ctx: Context for the operation
//   - reaction: The reaction object to add
//
// Returns:
//   - error: Any error that occurred during the operation
func (r *mongoReactionRepository) AddReaction(ctx context.Context, reaction *domain.Reaction) error {
	// Use a session to ensure transactional consistency
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Transactional logic
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check if reaction already exists
		filter := bson.M{
			"userid":      reaction.UserID,
			"contentid":   reaction.ContentID,
			"reactiontype": reaction.ReactionType,
		}
		
		var existingReaction domain.Reaction
		err := r.collection.FindOne(sessCtx, filter).Decode(&existingReaction)
		if err == nil {
			return err // Error or reaction already exists
		}

		// Insert the new reaction
		_, err = r.collection.InsertOne(sessCtx, reaction)
		if err != nil {
			return nil, err
		}

		// Increment the correct counter on the post
		updateFilter := bson.M{"_id": reaction.ContentID}
		update := bson.M{"$inc": bson.M{fmt.Sprintf("reactioncounts.%s", reaction.ReactionType): 1}}
		_, err = r.collection.Database().Collection("posts").UpdateOne(sessCtx, updateFilter, update)
		
		return nil, err
	})

	return err
}

// RemoveReaction deletes an existing reaction from the database and updates the content's reaction counters.
// This operation is performed within a transaction to ensure data consistency.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user who created the reaction
//   - contentID: ID of the content item
//   - reactionType: Type of reaction to remove
//
// Returns:
//   - error: Any error that occurred during the operation
func (r *mongoReactionRepository) RemoveReaction(ctx context.Context, userID, contentID primitive.ObjectID, reactionType domain.ReactionType) error {
	// Use a session to ensure transactional consistency
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Transactional logic
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete the reaction
		filter := bson.M{
			"userid":      userID,
			"contentid":   contentID,
			"reactiontype": reactionType,
		}
		
		result, err := r.collection.DeleteOne(sessCtx, filter)
		if err != nil || result.DeletedCount == 0 {
			return err // Error or reaction didn't exist
		}

		// Decrement the correct counter on the post
		updateFilter := bson.M{"_id": contentID}
		update := bson.M{"$inc": bson.M{fmt.Sprintf("reactioncounts.%s", reactionType): -1}}
		_, err = r.collection.Database().Collection("posts").UpdateOne(sessCtx, updateFilter, update)
		
		return nil, err
	})

	return err
}
