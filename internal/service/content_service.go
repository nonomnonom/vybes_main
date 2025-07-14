package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"
	"vybes/internal/config"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/media"
	"vybes/pkg/storage"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentService defines the interface for content business logic.
type ContentService interface {
	CreatePost(ctx context.Context, userID, caption string, visibility domain.PostVisibility, fileHeader *multipart.FileHeader) (*domain.Post, error)
	Repost(ctx context.Context, userID, originalPostID string) (*domain.Post, error)
	GetRepostsByUser(ctx context.Context, userID string, limit int) ([]domain.Post, error)
	CreateComment(ctx context.Context, userID, postID, text string) (*domain.Comment, error)
	GetComments(ctx context.Context, postID string, limit int) ([]domain.Comment, error)
	RecordView(ctx context.Context, postID string) error
}

type contentService struct {
	contentRepo      repository.ContentRepository
	userRepo         repository.UserRepository
	storage          storage.Client
	notificationPublisher NotificationPublisher
	mediaProcessor   media.Processor
	cfg              *config.Config
}

// NewContentService creates a new content service.
func NewContentService(contentRepo repository.ContentRepository, userRepo repository.UserRepository, storage storage.Client, notificationPublisher NotificationPublisher, mediaProcessor media.Processor, cfg *config.Config) ContentService {
	return &contentService{
		contentRepo:      contentRepo,
		userRepo:         userRepo,
		storage:          storage,
		notificationPublisher: notificationPublisher,
		mediaProcessor:   mediaProcessor,
		cfg:              cfg,
	}
}

func (s *contentService) CreatePost(ctx context.Context, userIDStr, caption string, visibility domain.PostVisibility, fileHeader *multipart.FileHeader) (*domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	// --- Start FFmpeg Processing ---
	// 1. Save the uploaded file temporarily
	tempFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if _, err = io.Copy(tempFile, src); err != nil {
		return nil, fmt.Errorf("failed to copy to temp file: %w", err)
	}
	tempFile.Close() // Close the file so ffmpeg can access it

	// 2. Generate thumbnail
	thumbnailPath := tempFile.Name() + ".jpg"
	defer os.Remove(thumbnailPath) // Clean up the thumbnail file

	if err := s.mediaProcessor.GenerateThumbnail(tempFile.Name(), thumbnailPath); err != nil {
		// We can decide if a post should fail if thumbnail generation fails.
		// For now, let's allow it but log the error.
		// log.Error().Err(err).Msg("Failed to generate thumbnail, proceeding without it")
		thumbnailPath = "" // No thumbnail
	}
	// --- End FFmpeg Processing ---

	// --- Start Uploading to MinIO ---
	// 3. Upload original video
	videoFile, err := os.Open(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to reopen temp video file: %w", err)
	}
	defer videoFile.Close()

	videoObjectName := fmt.Sprintf("posts/%s/%s.mp4", userID.Hex(), uuid.New().String())
	videoUploadInfo, err := s.storage.UploadFile(ctx, s.cfg.MinioBucketName, videoObjectName, videoFile, fileHeader.Size, "video/mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// 4. Upload thumbnail
	var thumbnailURL string
	if thumbnailPath != "" {
		thumbFile, err := os.Open(thumbnailPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open thumbnail file: %w", err)
		}
		defer thumbFile.Close()

		thumbStat, _ := thumbFile.Stat()
		thumbObjectName := fmt.Sprintf("posts/%s/%s.jpg", userID.Hex(), uuid.New().String())
		thumbUploadInfo, err := s.storage.UploadFile(ctx, s.cfg.MinioBucketName, thumbObjectName, thumbFile, thumbStat.Size(), "image/jpeg")
		if err != nil {
			// Log error but don't fail the post
		} else {
			thumbnailURL = thumbUploadInfo.Location
		}
	}
	// --- End Uploading to MinIO ---

	// 5. Create database documents
	content := &domain.Content{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		Type:         domain.ContentTypeVideo,
		URL:          videoUploadInfo.Location,
		ThumbnailURL: thumbnailURL,
		Caption:      caption,
		CreatedAt:    time.Now(),
	}
	if err := s.contentRepo.CreateContent(ctx, content); err != nil {
		return nil, err
	}

	post := &domain.Post{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		ContentID:  content.ID,
		Visibility: visibility,
		CreatedAt:  time.Now(),
	}
	if err := s.contentRepo.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	if err := s.userRepo.IncrementPostCount(ctx, userID, 1); err != nil {
		// Log this error
	}

	return post, nil
}

func (s *contentService) Repost(ctx context.Context, userIDStr, originalPostIDStr string) (*domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}
	originalPostID, err := primitive.ObjectIDFromHex(originalPostIDStr)
	if err != nil {
		return nil, err
	}

	originalPost, err := s.contentRepo.GetPostByID(ctx, originalPostID)
	if err != nil {
		return nil, err
	}
	if originalPost == nil {
		return nil, errors.New("original post not found")
	}

	// Reposts inherit the visibility of the original post.
	repost := &domain.Post{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		ContentID:      originalPost.ContentID,
		OriginalPostID: &originalPost.ID,
		Visibility:     originalPost.Visibility,
		CreatedAt:      time.Now(),
	}
	if err := s.contentRepo.CreatePost(ctx, repost); err != nil {
		return nil, err
	}

	if err := s.contentRepo.IncrementRepostCount(ctx, originalPost.ID); err != nil {
		// Log this error
	}
	
	if err := s.userRepo.IncrementPostCount(ctx, userID, 1); err != nil {
		// Log this error
	}

	return repost, nil
}

func (s *contentService) GetRepostsByUser(ctx context.Context, userIDStr string, limit int) ([]domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}
	return s.contentRepo.GetRepostsByUser(ctx, userID, limit)
}

func (s *contentService) CreateComment(ctx context.Context, userIDStr, postIDStr, text string) (*domain.Comment, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}

	post, err := s.contentRepo.GetPostByID(ctx, postID)
	if err != nil || post == nil {
		return nil, err
	}

	comment := &domain.Comment{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		PostID:    postID,
		Text:      text,
		CreatedAt: time.Now(),
	}
	if err := s.contentRepo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	if err := s.contentRepo.IncrementCommentCount(ctx, postID); err != nil {
		return nil, err
	}

	// Publish notification event
	go s.notificationPublisher.Publish(domain.Notification{
		UserID:  post.UserID, // The post author receives the notification
		ActorID: userID,
		Type:    domain.NotificationTypeComment,
		PostID:  &post.ID,
	})

	return comment, nil
}

func (s *contentService) GetComments(ctx context.Context, postIDStr string, limit int) ([]domain.Comment, error) {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}
	return s.contentRepo.GetCommentsByPost(ctx, postID, limit)
}

func (s *contentService) RecordView(ctx context.Context, postIDStr string) error {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}
	return s.contentRepo.IncrementViewCount(ctx, postID)
}