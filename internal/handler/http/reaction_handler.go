package http

import (
	"net/http"
	"vybes/internal/domain"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// ReactionHandler handles HTTP requests for reactions.
type ReactionHandler struct {
	reactionService service.ReactionService
}

// NewReactionHandler creates a new ReactionHandler.
func NewReactionHandler(reactionService service.ReactionService) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
	}
}

// AddLike is the handler for liking a post.
func (h *ReactionHandler) AddLike(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")

	err := h.reactionService.AddReaction(c.Request.Context(), userID.(string), postID, string(domain.ReactionTypeLike))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post liked successfully"})
}

// RemoveLike is the handler for unliking a post.
func (h *ReactionHandler) RemoveLike(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")

	err := h.reactionService.RemoveReaction(c.Request.Context(), userID.(string), postID, string(domain.ReactionTypeLike))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post unliked successfully"})
}
