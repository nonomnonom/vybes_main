package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CounterRepository defines the interface for counter operations.
type CounterRepository interface {
	GetNextSequence(ctx context.Context, name string) (int64, error)
}

type mongoCounterRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoCounterRepository creates a new counter repository.
func NewMongoCounterRepository(db *mongo.Database) CounterRepository {
	return &mongoCounterRepository{
		db:         db,
		collection: "counters",
	}
}

// GetNextSequence atomically finds and increments a counter.
func (r *mongoCounterRepository) GetNextSequence(ctx context.Context, name string) (int64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter domain.Counter
	err := r.db.Collection(r.collection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.SequenceValue, nil
}
