package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"vybes/internal/domain"
)

// SessionRepositoryInterface defines the interface for session data operations.
type SessionRepositoryInterface interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	Block(ctx context.Context, id primitive.ObjectID) error
}
const sessionCollection = "sessions"

// Ensure *SessionRepository implements SessionRepositoryInterface
var _ SessionRepositoryInterface = (*SessionRepository)(nil)

type SessionRepository struct {
	db *mongo.Database
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	_, err := r.db.Collection(sessionCollection).InsertOne(ctx, session)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Collection(sessionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Collection(sessionCollection).FindOne(ctx, bson.M{"refresh_token": refreshToken}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Block(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(sessionCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_blocked": true}},
	)
	return err
}
