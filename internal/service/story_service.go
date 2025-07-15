package service

import (
	"context"
	"fmt"

	"mime/multipart"
	"time"
	"vybes/internal/config"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/storage"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StoryService defines the interface for story business logic.
type StoryService interface {
	CreateStory(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (*domain.Story, error)
	GetStoryFeed(ctx context.Context, userID string) ([]domain.Story, error)
}

type storyService struct {
	storyRepo  repository.StoryRepository
	followRepo repository.FollowRepository
	storage    storage.Client
	cfg        *config.Config
}

// NewStoryService creates a new story service.
func NewStoryService(storyRepo repository.StoryRepository, followRepo repository.FollowRepository, storage storage.Client, cfg *config.Config) StoryService {
	return &storyService{
		storyRepo:  storyRepo,
		followRepo: followRepo,
		storage:    storage,
		cfg:        cfg,
	}
}

func (s *storyService) CreateStory(ctx context.Context, userIDStr string, fileHeader *multipart.FileHeader) (*domain.Story, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Generate a unique object name
	objectName := fmt.Sprintf("stories/%s/%s", userID.Hex(), uuid.New().String())
	contentType := fileHeader.Header.Get("Content-Type")

	// Upload to R2
	uploadInfo, err := s.storage.UploadFile(ctx, s.cfg.R2StoriesBucket, objectName, file, fileHeader.Size, contentType)
	if err != nil {
		return nil, err
	}

	// Create story metadata in MongoDB
	story := &domain.Story{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MediaURL:  uploadInfo.Location, // Or construct a public URL
		MediaType: contentType,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.storyRepo.Create(ctx, story); err != nil {
		// TODO: Implement logic to delete the object from R2 if this fails
		return nil, err
	}

	return story, nil
}

func (s *storyService) GetStoryFeed(ctx context.Context, userIDStr string) ([]domain.Story, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	// Get IDs of users the current user is following
	followingIDs, err := s.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Also include the user's own stories in their feed
	followingIDs = append(followingIDs, userID)

	return s.storyRepo.GetStoriesByUsers(ctx, followingIDs)
}
