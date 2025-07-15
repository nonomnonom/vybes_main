package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"vybes/internal/domain"
	"vybes/internal/repository"
)

// SessionServiceInterface defines the interface for session business logic.
type SessionServiceInterface interface {
	Create(ctx context.Context, userID primitive.ObjectID, refreshToken, userAgent, clientIP string) (*domain.Session, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error)
	Block(ctx context.Context, id primitive.ObjectID) error
}
// Ensure *SessionService implements SessionServiceInterface
var _ SessionServiceInterface = (*SessionService)(nil)

type SessionService struct {
	sessionRepo *repository.SessionRepository
}

func NewSessionService(sessionRepo *repository.SessionRepository) *SessionService {
	return &SessionService{sessionRepo: sessionRepo}
}

func (s *SessionService) Create(ctx context.Context, userID primitive.ObjectID, refreshToken, userAgent, clientIP string) (*domain.Session, error) {
	session := &domain.Session{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7), // 7 days
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Session, error) {
	return s.sessionRepo.GetByID(ctx, id)
}

func (s *SessionService) Block(ctx context.Context, id primitive.ObjectID) error {
	return s.sessionRepo.Block(ctx, id)
}
