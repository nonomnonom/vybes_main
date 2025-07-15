package http

import (
	"net/http"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// StoryHandler handles HTTP requests for stories.
type StoryHandler struct {
	storyService service.StoryService
}

// NewStoryHandler creates a new StoryHandler.
func NewStoryHandler(storyService service.StoryService) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
	}
}

// CreateStory is the handler for uploading a new story.
func (h *StoryHandler) CreateStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	file, err := c.FormFile("media")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media file is required"})
		return
	}

	story, err := h.storyService.CreateStory(c.Request.Context(), userID.(string), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	c.JSON(http.StatusCreated, story)
}

// GetStoryFeed is the handler for getting the user's story feed.
func (h *StoryHandler) GetStoryFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	feed, err := h.storyService.GetStoryFeed(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get story feed"})
		return
	}

	c.JSON(http.StatusOK, feed)
}
