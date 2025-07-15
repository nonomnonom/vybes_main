package service

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ISessionService defines the interface for session business logic.
type ISessionService interface {
	Create(ctx context.Context, userID primitive.ObjectID, refreshToken, userAgent, clientIP string) (*domain.Session, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error)
	Block(ctx context.Context, id primitive.ObjectID) error
}

// Ensure *SessionService implements ISessionService
var _ ISessionService = (*SessionService)(nil)