package http

import (
	"net/http"
	"strconv"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// FeedHandler handles HTTP requests for feeds.
type FeedHandler struct {
	feedService service.FeedService
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(feedService service.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
	}
}

// GetForYouFeed is the handler for getting the user's "For You" feed.
func (h *FeedHandler) GetForYouFeed(c *gin.Context) {
	userID, _ := c.Get("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	feed, err := h.feedService.GetForYouFeed(c.Request.Context(), userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feed"})
		return
	}
	c.JSON(http.StatusOK, feed)
}

// GetFriendFeed is the handler for getting the user's "Friend" feed.
func (h *FeedHandler) GetFriendFeed(c *gin.Context) {
	userID, _ := c.Get("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	feed, err := h.feedService.GetFriendFeed(c.Request.Context(), userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get friend feed"})
		return
	}
	c.JSON(http.StatusOK, feed)
}