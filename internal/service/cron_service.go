package service

import (
	"context"
	"strings"
	"vybes/internal/config"
	"vybes/internal/repository"
	"vybes/pkg/storage"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CronService manages scheduled tasks.
type CronService struct {
	cfg       *config.Config
	storyRepo repository.StoryRepository
	storage   storage.Client
}

// NewCronService creates a new cron service.
func NewCronService(cfg *config.Config, storyRepo repository.StoryRepository, storage storage.Client) *CronService {
	return &CronService{
		cfg:       cfg,
		storyRepo: storyRepo,
		storage:   storage,
	}
}

// Start initializes and starts the cron jobs.
func (s *CronService) Start() {
	c := cron.New()

	// Schedule a job to run every hour to clean up expired stories.
	c.AddFunc("@hourly", s.cleanupExpiredStories)

	log.Info().Msg("Starting cron jobs...")
	c.Start()
}

func (s *CronService) cleanupExpiredStories() {
	log.Info().Msg("Running expired stories cleanup job...")
	ctx := context.Background()

	expiredStories, err := s.storyRepo.FindExpired(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find expired stories")
		return
	}

	if len(expiredStories) == 0 {
		log.Info().Msg("No expired stories to clean up.")
		return
	}

	var storyIDsToDelete []primitive.ObjectID
	var objectsToDelete []string

	for _, story := range expiredStories {
		storyIDsToDelete = append(storyIDsToDelete, story.ID)
		// Extract object name from URL
		// This logic is brittle. A better way is to store the object name directly in the story document.
		// For now, we'll stick with this.
		objectName := strings.TrimPrefix(story.MediaURL, s.cfg.R2Endpoint+"/"+s.cfg.R2StoriesBucket+"/")
		if objectName != story.MediaURL { // Ensure prefix was actually trimmed
			objectsToDelete = append(objectsToDelete, objectName)
		}
	}

	// Delete files from R2
	for _, objectName := range objectsToDelete {
		err := s.storage.DeleteFile(ctx, s.cfg.R2StoriesBucket, objectName)
		if err != nil {
			log.Error().Err(err).Str("object", objectName).Msg("Failed to delete story media from storage")
			// We continue even if one fails, to attempt deleting others.
		}
	}

	// Delete metadata from MongoDB
	err = s.storyRepo.DeleteMany(ctx, storyIDsToDelete)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete expired story metadata from DB")
		return
	}

	log.Info().Int("count", len(storyIDsToDelete)).Msg("Successfully cleaned up expired stories.")
}
