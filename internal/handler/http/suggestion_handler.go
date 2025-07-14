package http

import (
	"net/http"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// SuggestionHandler handles HTTP requests for user suggestions.
type SuggestionHandler struct {
	suggestionService service.SuggestionService
}

// NewSuggestionHandler creates a new SuggestionHandler.
func NewSuggestionHandler(suggestionService service.SuggestionService) *SuggestionHandler {
	return &SuggestionHandler{
		suggestionService: suggestionService,
	}
}

// GetSuggestions is the handler for getting user follow suggestions.
func (h *SuggestionHandler) GetSuggestions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	suggestions, err := h.suggestionService.GetSuggestions(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get suggestions"})
		return
	}

	c.JSON(http.StatusOK, suggestions)
}