package service

import (
	"context"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionService defines the interface for reaction business logic.
type ReactionService interface {
	AddReaction(ctx context.Context, userID, postID, reactionType string) error
	RemoveReaction(ctx context.Context, userID, postID, reactionType string) error
}

type reactionService struct {
	reactionRepo          repository.ReactionRepository
	contentRepo           repository.ContentRepository
	userRepo              repository.UserRepository
	notificationPublisher NotificationPublisher
}

// NewReactionService creates a new reaction service.
func NewReactionService(reactionRepo repository.ReactionRepository, contentRepo repository.ContentRepository, userRepo repository.UserRepository, notificationPublisher NotificationPublisher) ReactionService {
	return &reactionService{
		reactionRepo:          reactionRepo,
		contentRepo:           contentRepo,
		userRepo:              userRepo,
		notificationPublisher: notificationPublisher,
	}
}

func (s *reactionService) AddReaction(ctx context.Context, userIDStr, postIDStr, reactionTypeStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return err
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}

	post, err := s.contentRepo.GetPostByID(ctx, postID)
	if err != nil || post == nil {
		return err
	}

	reaction := &domain.Reaction{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		PostID:    postID,
		Type:      domain.ReactionType(reactionTypeStr),
		CreatedAt: time.Now(),
	}

	if err := s.reactionRepo.AddReaction(ctx, reaction); err != nil {
		return err
	}
	if reaction.Type == domain.ReactionTypeLike {
		if err := s.userRepo.IncrementTotalLikes(ctx, post.UserID, 1); err != nil {
			return err
		}
		// Publish notification event
		go s.notificationPublisher.Publish(domain.Notification{
			UserID:  post.UserID, // The post author receives the notification
			ActorID: userID,
			Type:    domain.NotificationTypeLike,
			PostID:  &post.ID,
		})
	}
	return nil
}

func (s *reactionService) RemoveReaction(ctx context.Context, userIDStr, postIDStr, reactionTypeStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return err
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}
	reactionType := domain.ReactionType(reactionTypeStr)

	post, err := s.contentRepo.GetPostByID(ctx, postID)
	if err != nil || post == nil {
		return err
	}

	if err := s.reactionRepo.RemoveReaction(ctx, userID, postID, reactionType); err != nil {
		return err
	}
	if reactionType == domain.ReactionTypeLike {
		if err := s.userRepo.IncrementTotalLikes(ctx, post.UserID, -1); err != nil {
			return err
		}
	}
	return nil
}
