package service

import (
	"context"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService defines the interface for notification business logic.
type NotificationService interface {
	CreateNotification(ctx context.Context, userID, actorID primitive.ObjectID, notifType domain.NotificationType, postID *primitive.ObjectID) error
	GetNotifications(ctx context.Context, userID string, page, limit int) ([]domain.Notification, error)
	MarkNotificationsAsRead(ctx context.Context, userID string, notificationIDs []string) (int64, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService creates a new notification service.
func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID, actorID primitive.ObjectID, notifType domain.NotificationType, postID *primitive.ObjectID) error {
	// Avoid self-notification
	if userID == actorID {
		return nil
	}

	notification := &domain.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ActorID:   actorID,
		Type:      notifType,
		PostID:    postID,
		Read:      false,
		CreatedAt: time.Now(),
	}
	return s.notificationRepo.CreateNotification(ctx, notification)
}

func (s *notificationService) GetNotifications(ctx context.Context, userIDStr string, page, limit int) ([]domain.Notification, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}
	return s.notificationRepo.GetUserNotifications(ctx, userID, page, limit)
}

func (s *notificationService) MarkNotificationsAsRead(ctx context.Context, userIDStr string, notificationIDStrs []string) (int64, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return 0, err
	}

	var successCount int64
	for _, idStr := range notificationIDStrs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err == nil {
			if err := s.notificationRepo.MarkAsRead(ctx, id, userID); err == nil {
				successCount++
			}
		}
	}

	return successCount, nil
}
