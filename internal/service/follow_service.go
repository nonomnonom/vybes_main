package service

import (
	"context"
	"errors"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FollowService defines the interface for follow business logic.
type FollowService interface {
	FollowUser(ctx context.Context, followerID, followingUsername string) error
	UnfollowUser(ctx context.Context, followerID, followingUsername string) error
}

type followService struct {
	followRepo            repository.FollowRepository
	userRepo              repository.UserRepository
	notificationPublisher NotificationPublisher
}

// NewFollowService creates a new follow service.
func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository, notificationPublisher NotificationPublisher) FollowService {
	return &followService{
		followRepo:            followRepo,
		userRepo:              userRepo,
		notificationPublisher: notificationPublisher,
	}
}

func (s *followService) FollowUser(ctx context.Context, followerIDStr, followingUsername string) error {
	followerID, err := primitive.ObjectIDFromHex(followerIDStr)
	if err != nil {
		return errors.New("invalid follower ID format")
	}

	userToFollow, err := s.userRepo.FindByUsername(ctx, followingUsername)
	if err != nil {
		return err
	}
	if userToFollow == nil {
		return errors.New("user to follow not found")
	}

	if followerID == userToFollow.ID {
		return errors.New("cannot follow yourself")
	}

	if err := s.followRepo.Follow(ctx, followerID, userToFollow.ID); err != nil {
		return err
	}

	// Publish notification event
	go s.notificationPublisher.Publish(domain.Notification{
		UserID:  userToFollow.ID, // The one being followed receives the notification
		ActorID: followerID,
		Type:    domain.NotificationTypeFollow,
	})

	return nil
}

func (s *followService) UnfollowUser(ctx context.Context, followerIDStr, followingUsername string) error {
	followerID, err := primitive.ObjectIDFromHex(followerIDStr)
	if err != nil {
		return errors.New("invalid follower ID format")
	}

	userToUnfollow, err := s.userRepo.FindByUsername(ctx, followingUsername)
	if err != nil {
		return err
	}
	if userToUnfollow == nil {
		return errors.New("user to unfollow not found")
	}

	return s.followRepo.Unfollow(ctx, followerID, userToUnfollow.ID)
}
