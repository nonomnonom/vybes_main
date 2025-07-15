package service

import (
	"context"
	"vybes/internal/domain"
	"vybes/internal/repository"
)

// SearchService defines the interface for search business logic.
type SearchService interface {
	SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error)
}

type searchService struct {
	userRepo repository.UserRepository
}

// NewSearchService creates a new search service.
func NewSearchService(userRepo repository.UserRepository) SearchService {
	return &searchService{
		userRepo: userRepo,
	}
}

func (s *searchService) SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error) {
	return s.userRepo.SearchUsers(ctx, query, limit)
}
