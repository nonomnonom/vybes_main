package http

import (
	"net/http"
	"strconv"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// BookmarkHandler handles HTTP requests for bookmarks.
type BookmarkHandler struct {
	bookmarkService service.BookmarkService
}

// NewBookmarkHandler creates a new BookmarkHandler.
func NewBookmarkHandler(bookmarkService service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkService: bookmarkService,
	}
}

// AddBookmark is the handler for adding a bookmark.
func (h *BookmarkHandler) AddBookmark(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")

	err := h.bookmarkService.AddBookmark(c.Request.Context(), userID.(string), postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post bookmarked successfully"})
}

// RemoveBookmark is the handler for removing a bookmark.
func (h *BookmarkHandler) RemoveBookmark(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")

	err := h.bookmarkService.RemoveBookmark(c.Request.Context(), userID.(string), postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bookmark removed successfully"})
}

// GetBookmarks is the handler for getting the user's bookmarked posts.
func (h *BookmarkHandler) GetBookmarks(c *gin.Context) {
	userID, _ := c.Get("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	posts, err := h.bookmarkService.GetBookmarks(c.Request.Context(), userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookmarks"})
		return
	}
	c.JSON(http.StatusOK, posts)
}
