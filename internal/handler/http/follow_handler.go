package http

import (
	"net/http"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// FollowHandler handles HTTP requests for follow relationships.
type FollowHandler struct {
	followService service.FollowService
}

// NewFollowHandler creates a new FollowHandler.
func NewFollowHandler(followService service.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// FollowUser is the handler for following a user.
func (h *FollowHandler) FollowUser(c *gin.Context) {
	followerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	usernameToFollow := c.Param("username")

	err := h.followService.FollowUser(c.Request.Context(), followerID.(string), usernameToFollow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}

// UnfollowUser is the handler for unfollowing a user.
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	followerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	usernameToUnfollow := c.Param("username")

	err := h.followService.UnfollowUser(c.Request.Context(), followerID.(string), usernameToUnfollow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}
