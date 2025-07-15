package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
	"vybes/internal/config"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/storage"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentService defines the interface for content business logic.
type ContentService interface {
	CreatePost(ctx context.Context, userID, caption string, visibility domain.PostVisibility, fileHeader *multipart.FileHeader) (*domain.Post, error)
	DeletePost(ctx context.Context, userID, postID string) error
	Repost(ctx context.Context, userID, originalPostID string) (*domain.Post, error)
	GetRepostsByUser(ctx context.Context, userID string, limit int) ([]domain.Post, error)
	CreateComment(ctx context.Context, userID, postID, text string) (*domain.Comment, error)
	GetComments(ctx context.Context, postID string, limit int) ([]domain.Comment, error)
	RecordView(ctx context.Context, postID string) error
}

type contentService struct {
	contentRepo           repository.ContentRepository
	userRepo              repository.UserRepository
	storage               storage.Client
	notificationPublisher NotificationPublisher
	cfg                   *config.Config
}

// NewContentService creates a new content service.
func NewContentService(contentRepo repository.ContentRepository, userRepo repository.UserRepository, storage storage.Client, notificationPublisher NotificationPublisher, cfg *config.Config) ContentService {
	return &contentService{
		contentRepo:           contentRepo,
		userRepo:              userRepo,
		storage:               storage,
		notificationPublisher: notificationPublisher,
		cfg:                   cfg,
	}
}

func (s *contentService) CreatePost(ctx context.Context, userIDStr, caption string, visibility domain.PostVisibility, fileHeader *multipart.FileHeader) (*domain.Post, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	objectName := fmt.Sprintf("posts/%s/%s", userID.Hex(), uuid.New().String())
	contentType := fileHeader.Header.Get("Content-Type")

	uploadInfo, err := s.storage.UploadFile(ctx, s.cfg.R2PostsBucket, objectName, file, fileHeader.Size, contentType)
	if err != nil {
		return nil, err
	}

	content := &domain.Content{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Type:      domain.ContentTypeVideo, // Assuming video for now
		URL:       uploadInfo.Location,
		Caption:   caption,
		CreatedAt: time.Now(),
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

func (s *contentService) DeletePost(ctx context.Context, userIDStr, postIDStr string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return errors.New("invalid post ID format")
	}

	// 1. Get the post to verify ownership and get content details
	post, err := s.contentRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}
	if post.UserID != userID {
		return errors.New("user not authorized to delete this post")
	}

	// 2. Get the associated content to find file URLs
	content, err := s.contentRepo.GetContentByID(ctx, post.ContentID)
	if err != nil {
		// Log this, but proceed with deletion of the post itself
	}

	// 3. Delete file from R2
	if content != nil && content.URL != "" {
		// Extract object name from URL (assuming URL format: https://bucket.r2.cloudflarestorage.com/object-name)
		objectName := strings.TrimPrefix(content.URL, s.cfg.R2Endpoint+"/"+s.cfg.R2PostsBucket+"/")
		s.storage.DeleteFile(ctx, s.cfg.R2PostsBucket, objectName)
	}

	// 4. Delete the post and content documents
	if err := s.contentRepo.DeletePost(ctx, postID); err != nil {
		return err
	}
	if err := s.contentRepo.DeleteContent(ctx, post.ContentID); err != nil {
		// Log this error
	}

	// TODO: Delete associated comments, likes, bookmarks, notifications in a transaction or background job.

	// 5. Decrement user's post count
	if err := s.userRepo.IncrementPostCount(ctx, userID, -1); err != nil {
		// Log this error
	}

	return nil
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
