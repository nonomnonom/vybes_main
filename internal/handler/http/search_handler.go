package http

import (
	"net/http"
	"strconv"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles HTTP requests for searching.
type SearchHandler struct {
	searchService service.SearchService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchUsers is the handler for searching users.
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, err := h.searchService.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
