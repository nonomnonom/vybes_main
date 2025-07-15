package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/storage"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentService defines the interface for content business logic operations.
// Content services handle the creation, management, and processing of posts,
// including file uploads, content validation, and user interaction tracking.
type ContentService interface {
	// CreatePost creates a new post with optional file upload
	CreatePost(ctx context.Context, userID primitive.ObjectID, caption string, file *multipart.FileHeader, visibility domain.PostVisibility) (*domain.Post, error)
	// GetPostByID retrieves a specific post by its ID
	GetPostByID(ctx context.Context, postID primitive.ObjectID) (*domain.Post, error)
	// GetPostsByUserID retrieves posts created by a specific user
	GetPostsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Post, error)
	// DeletePost removes a post and its associated content
	DeletePost(ctx context.Context, postID, userID primitive.ObjectID) error
	// GetFeedPosts retrieves posts for a user's feed based on followed users
	GetFeedPosts(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Post, error)
}

// contentService implements ContentService with business logic for content management
type contentService struct {
	contentRepository     repository.ContentRepository
	userRepository        repository.UserRepository
	storageClient         storage.Client
	notificationPublisher NotificationPublisher
	config                *domain.Config
}

// NewContentService creates a new content service instance with all required dependencies.
// The service handles content creation, file uploads, and post management operations.
//
// Parameters:
//   - contentRepository: Repository for content data operations
//   - userRepository: Repository for user data operations
//   - storageClient: Client for file storage operations
//   - notificationPublisher: Publisher for real-time notifications
//   - config: Application configuration
//
// Returns:
//   - ContentService: A configured content service ready for use
func NewContentService(contentRepository repository.ContentRepository, userRepository repository.UserRepository, storageClient storage.Client, notificationPublisher NotificationPublisher, config *domain.Config) ContentService {
	return &contentService{
		contentRepository:     contentRepository,
		userRepository:        userRepository,
		storageClient:         storageClient,
		notificationPublisher: notificationPublisher,
		config:                config,
	}
}

// CreatePost creates a new post with optional file upload and publishes notifications.
// The method handles file validation, upload to cloud storage, and post creation
// with proper error handling and cleanup on failure.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user creating the post
//   - caption: Text caption for the post
//   - file: Optional file to upload with the post
//   - visibility: Post visibility setting (public, private, followers)
//
// Returns:
//   - *domain.Post: The created post with all metadata
//   - error: Any error that occurred during post creation
func (s *contentService) CreatePost(ctx context.Context, userID primitive.ObjectID, caption string, file *multipart.FileHeader, visibility domain.PostVisibility) (*domain.Post, error) {
	// Validate user exists
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	var contentURL string
	var contentType domain.ContentType

	// Handle file upload if provided
	if file != nil {
		// Validate file type and size
		if err := s.validateFile(file); err != nil {
			return nil, fmt.Errorf("file validation failed: %w", err)
		}

		// Upload file to cloud storage
		contentURL, err = s.uploadFile(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("file upload failed: %w", err)
		}

		// Determine content type based on file extension
		contentType = s.determineContentType(file.Filename)
	} else {
		contentType = domain.ContentTypeVideo // Assuming video for now
	}

	// Create post object
	post := &domain.Post{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		Caption:    caption,
		ContentURL: contentURL,
		Type:       contentType,
		Visibility: visibility,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save post to database
	if err := s.contentRepository.CreatePost(ctx, post); err != nil {
		// Log this error
		log.Error().Err(err).Msg("Failed to save post to database")
		
		// Clean up uploaded file if post creation failed
		if contentURL != "" {
			if deleteErr := s.storageClient.DeleteFile(ctx, s.config.R2BucketName, contentURL); deleteErr != nil {
				log.Error().Err(deleteErr).Msg("Failed to delete uploaded file after post creation failure")
			}
		}
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Publish notification for new post (if public)
	if visibility == domain.PostVisibilityPublic {
		// TODO: Implement notification publishing for new posts
	}

	return post, nil
}

// DeletePost removes a post and its associated content from the system.
// This operation includes file cleanup from cloud storage and cascading
// deletion of related data like comments and reactions.
//
// Parameters:
//   - ctx: Context for the operation
//   - postID: ID of the post to delete
//   - userID: ID of the user requesting deletion (for authorization)
//
// Returns:
//   - error: Any error that occurred during post deletion
func (s *contentService) DeletePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	// 1. Get the post to verify ownership and get content details
	post, err := s.contentRepository.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return fmt.Errorf("post not found")
	}

	// Verify user owns the post
	if post.UserID != userID {
		return fmt.Errorf("unauthorized: user does not own this post")
	}

	// 2. Get the associated content to find file URLs
	content, err := s.contentRepository.GetContentByID(ctx, post.ContentID)
	if err != nil {
		// Log this, but proceed with deletion of the post itself
		log.Warn().Err(err).Msg("Failed to get content for post deletion")
	}

	// 3. Delete file from R2 if content exists
	if content != nil && content.FileURL != "" {
		// Extract object name from URL (assuming URL format: https://bucket.r2.cloudflarestorage.com/object-name)
		objectName := strings.TrimPrefix(content.FileURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", s.config.R2BucketName))
		
		if err := s.storageClient.DeleteFile(ctx, s.config.R2BucketName, objectName); err != nil {
			log.Error().Err(err).Msg("Failed to delete file from storage during post deletion")
		}
	}

	// 4. Delete the post and content documents
	if err := s.contentRepository.DeletePost(ctx, postID, userID); err != nil {
		// Log this error
		log.Error().Err(err).Msg("Failed to delete post from database")
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// TODO: Delete associated comments, likes, bookmarks, notifications in a transaction or background job.
	// This should be implemented as a background job to avoid blocking the user request.

	return nil
}
