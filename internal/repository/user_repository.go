package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByVID(ctx context.Context, vid int64) (*domain.User, error)
	FindManyByIDs(ctx context.Context, ids []primitive.ObjectID) ([]domain.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	IncrementTotalLikes(ctx context.Context, userID primitive.ObjectID, amount int) error
	IncrementPostCount(ctx context.Context, userID primitive.ObjectID, amount int) error
}

type mongoUserRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoUserRepository creates a new user repository with MongoDB.
func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		db:         db,
		collection: "users",
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Collection(r.collection).InsertOne(ctx, user)
	return err
}

func (r *mongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) FindByVID(ctx context.Context, vid int64) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection(r.collection).FindOne(ctx, bson.M{"vid": vid}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) FindManyByIDs(ctx context.Context, ids []primitive.ObjectID) ([]domain.User, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *mongoUserRepository) SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error) {
	filter := bson.M{"$text": bson.M{"$search": query}}
	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.db.Collection(r.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *mongoUserRepository) IncrementTotalLikes(ctx context.Context, userID primitive.ObjectID, amount int) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$inc": bson.M{"totallikecount": amount}}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *mongoUserRepository) IncrementPostCount(ctx context.Context, userID primitive.ObjectID, amount int) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$inc": bson.M{"postcount": amount}}
	_, err := r.db.Collection(r.collection).UpdateOne(ctx, filter, update)
	return err
}