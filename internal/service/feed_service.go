package service

import (
	"context"
	"sort"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedService defines the interface for feed generation logic.
type FeedService interface {
	GetForYouFeed(ctx context.Context, userID string, limit int) ([]domain.Post, error)
	GetFriendFeed(ctx context.Context, userID string, limit int) ([]domain.Post, error)
}

type feedService struct {
	contentRepo repository.ContentRepository
	followRepo  repository.FollowRepository
}

// NewFeedService creates a new feed service.
func NewFeedService(contentRepo repository.ContentRepository, followRepo repository.FollowRepository) FeedService {
	return &feedService{
		contentRepo: contentRepo,
		followRepo:  followRepo,
	}
}

// GetForYouFeed fetches posts for the "For You" feed.
// This includes:
// 1. Posts from users the current user follows (visibility: public, friends).
// 2. The current user's own posts (visibility: public, friends, private).
// The results are combined, sorted by creation date, and limited.
func (s *feedService) GetForYouFeed(ctx context.Context, userIDStr string, limit int) ([]domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	followingIDs, err := s.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Fetch posts from followed users (public and friends-only)
	followedPosts, err := s.contentRepo.GetPostsByUsersWithVisibility(
		ctx,
		followingIDs,
		[]domain.PostVisibility{domain.VisibilityPublic, domain.VisibilityFriends},
		limit, // We fetch `limit` for each part, then sort and re-limit. Not perfectly efficient but works.
	)
	if err != nil {
		return nil, err
	}

	// 2. Fetch user's own posts (all visibilities)
	myPosts, err := s.contentRepo.GetPostsByUsersWithVisibility(
		ctx,
		[]primitive.ObjectID{userID},
		[]domain.PostVisibility{domain.VisibilityPublic, domain.VisibilityFriends, domain.VisibilityPrivate},
		limit,
	)
	if err != nil {
		return nil, err
	}

	// 3. Combine and sort
	combinedPosts := append(followedPosts, myPosts...)
	sort.Slice(combinedPosts, func(i, j int) bool {
		return combinedPosts[i].CreatedAt.After(combinedPosts[j].CreatedAt)
	})

	// 4. Apply limit
	if len(combinedPosts) > limit {
		return combinedPosts[:limit], nil
	}

	return combinedPosts, nil
}

// GetFriendFeed implements a feed of posts only from mutual follows (friends).
// It should only show posts with 'public' or 'friends' visibility.
func (s *feedService) GetFriendFeed(ctx context.Context, userIDStr string, limit int) ([]domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	followingIDs, err := s.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	followerIDs, err := s.followRepo.GetFollowerIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	followingSet := make(map[primitive.ObjectID]struct{})
	for _, id := range followingIDs {
		followingSet[id] = struct{}{}
	}

	var mutualIDs []primitive.ObjectID
	for _, id := range followerIDs {
		if _, found := followingSet[id]; found {
			mutualIDs = append(mutualIDs, id)
		}
	}

	if len(mutualIDs) == 0 {
		return []domain.Post{}, nil
	}

	// Friends can see public and friends-only posts.
	visibilities := []domain.PostVisibility{domain.VisibilityPublic, domain.VisibilityFriends}
	return s.contentRepo.GetPostsByUsersWithVisibility(ctx, mutualIDs, visibilities, limit)
}