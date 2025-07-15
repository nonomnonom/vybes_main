package service

import (
	"context"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookmarkService defines the interface for bookmark business logic.
type BookmarkService interface {
	AddBookmark(ctx context.Context, userID, postID string) error
	RemoveBookmark(ctx context.Context, userID, postID string) error
	GetBookmarks(ctx context.Context, userID string, page, limit int) ([]domain.Post, error)
}

type bookmarkService struct {
	bookmarkRepo repository.BookmarkRepository
	contentRepo  repository.ContentRepository
}

// NewBookmarkService creates a new bookmark service.
func NewBookmarkService(bookmarkRepo repository.BookmarkRepository, contentRepo repository.ContentRepository) BookmarkService {
	return &bookmarkService{
		bookmarkRepo: bookmarkRepo,
		contentRepo:  contentRepo,
	}
}

func (s *bookmarkService) AddBookmark(ctx context.Context, userIDStr, postIDStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return err
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}
	bookmark := &domain.Bookmark{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		PostID:    postID,
		CreatedAt: time.Now(),
	}
	return s.bookmarkRepo.CreateBookmark(ctx, bookmark)
}

func (s *bookmarkService) RemoveBookmark(ctx context.Context, userIDStr, postIDStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return err
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}
	return s.bookmarkRepo.DeleteBookmark(ctx, userID, postID)
}

func (s *bookmarkService) GetBookmarks(ctx context.Context, userIDStr string, page, limit int) ([]domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	bookmarks, err := s.bookmarkRepo.GetUserBookmarks(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	if len(bookmarks) == 0 {
		return []domain.Post{}, nil
	}

	// Extract post IDs from bookmarks
	var postIDs []primitive.ObjectID
	for _, b := range bookmarks {
		postIDs = append(postIDs, b.PostID)
	}

	// Fetch the full post details for the bookmarked posts
	return s.contentRepo.GetPostsByIDs(ctx, postIDs)
}
