package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ISessionRepository defines the interface for session data operations.
type ISessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
	Block(ctx context.Context, id primitive.ObjectID) error
}

// Ensure *SessionRepository implements ISessionRepository
var _ ISessionRepository = (*SessionRepository)(nil)