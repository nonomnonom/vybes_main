package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CounterRepository defines the interface for counter operations.
// Counters are used to generate unique sequential IDs for various entities
// like posts, comments, and other content types.
type CounterRepository interface {
	// GetNextSequence atomically increments and returns the next sequence number
	// for the specified counter name. This ensures unique ID generation.
	GetNextSequence(ctx context.Context, counterName string) (int64, error)
}

// mongoCounterRepository implements CounterRepository using MongoDB as the backend
type mongoCounterRepository struct {
	collection *mongo.Collection
}

// NewMongoCounterRepository creates a new counter repository instance with MongoDB backend.
// The repository handles atomic counter operations for generating unique sequential IDs.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - CounterRepository: A configured counter repository ready for use
func NewMongoCounterRepository(db *mongo.Database) CounterRepository {
	return &mongoCounterRepository{
		collection: db.Collection("counters"),
	}
}

// GetNextSequence atomically finds and increments a counter document in MongoDB.
// This operation is atomic and thread-safe, ensuring unique sequence generation
// even under concurrent access.
//
// Parameters:
//   - ctx: Context for the operation
//   - counterName: Name of the counter to increment (e.g., "posts", "comments")
//
// Returns:
//   - int64: The next sequence number
//   - error: Any error that occurred during the operation
func (r *mongoCounterRepository) GetNextSequence(ctx context.Context, counterName string) (int64, error) {
	// Use findOneAndUpdate with upsert to atomically increment the counter
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	
	var counter domain.Counter
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": counterName},
		bson.M{"$inc": bson.M{"seq": 1}},
		opts,
	).Decode(&counter)
	
	if err != nil {
		return 0, err
	}
	
	return counter.Seq, nil
}
