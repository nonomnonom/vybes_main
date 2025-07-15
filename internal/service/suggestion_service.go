package service

import (
	"context"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuggestionService defines the interface for user suggestion logic.
type SuggestionService interface {
	GetSuggestions(ctx context.Context, userID string) ([]domain.User, error)
}

type suggestionService struct {
	userRepo   repository.UserRepository
	followRepo repository.FollowRepository
}

// NewSuggestionService creates a new suggestion service.
func NewSuggestionService(userRepo repository.UserRepository, followRepo repository.FollowRepository) SuggestionService {
	return &suggestionService{
		userRepo:   userRepo,
		followRepo: followRepo,
	}
}

// GetSuggestions implements a "friends of friends" suggestion logic.
func (s *suggestionService) GetSuggestions(ctx context.Context, userIDStr string) ([]domain.User, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	// 1. Get the list of users the current user is following.
	followingIDs, err := s.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. For each user they are following, get the list of users *they* are following.
	suggestionMap := make(map[primitive.ObjectID]bool)
	for _, friendID := range followingIDs {
		friendsOfFriendIDs, err := s.followRepo.GetFollowingIDs(ctx, friendID)
		if err != nil {
			continue // Skip if there's an error for one user
		}
		for _, fofID := range friendsOfFriendIDs {
			// Don't suggest the user themselves or people they already follow.
			isFollowing := false
			for _, id := range followingIDs {
				if id == fofID {
					isFollowing = true
					break
				}
			}
			if fofID != userID && !isFollowing {
				suggestionMap[fofID] = true
			}
		}
	}

	// 3. Collect the unique IDs and fetch their user details.
	var suggestionIDs []primitive.ObjectID
	for id := range suggestionMap {
		suggestionIDs = append(suggestionIDs, id)
	}

	if len(suggestionIDs) == 0 {
		return []domain.User{}, nil
	}

	return s.userRepo.FindManyByIDs(ctx, suggestionIDs)
}
