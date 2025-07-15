package http

import (
	"net/http"
	"strconv"
	"vybes/internal/domain"
	"vybes/internal/service"

	"github.com/gin-gonic/gin"
)

// ContentHandler handles HTTP requests for content and comments.
type ContentHandler struct {
	contentService service.ContentService
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(contentService service.ContentService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}

// CreatePost is the handler for uploading new content as a post.
func (h *ContentHandler) CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	caption := c.PostForm("caption")
	visibilityStr := c.DefaultPostForm("visibility", string(domain.VisibilityPublic))
	visibility := domain.PostVisibility(visibilityStr)

	// Validate visibility
	switch visibility {
	case domain.VisibilityPublic, domain.VisibilityFriends, domain.VisibilityPrivate:
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visibility value"})
		return
	}

	file, err := c.FormFile("media")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media file is required"})
		return
	}

	post, err := h.contentService.CreatePost(c.Request.Context(), userID.(string), caption, visibility, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, post)
}

// DeletePost is the handler for deleting a post.
func (h *ContentHandler) DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")

	if err := h.contentService.DeletePost(c.Request.Context(), userID.(string), postID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Repost is the handler for reposting another post.
func (h *ContentHandler) Repost(c *gin.Context) {
	userID, _ := c.Get("userID")
	originalPostID := c.Param("postID")

	repost, err := h.contentService.Repost(c.Request.Context(), userID.(string), originalPostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, repost)
}

// GetRepostsByUser is the handler for getting a user's reposts.
func (h *ContentHandler) GetRepostsByUser(c *gin.Context) {
	userID := c.Param("userID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	posts, err := h.contentService.GetRepostsByUser(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reposts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// CreateComment is the handler for adding a comment to a post.
func (h *ContentHandler) CreateComment(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("postID")
	var request struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.contentService.CreateComment(c.Request.Context(), userID.(string), postID, request.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// GetComments is the handler for getting comments for a post.
func (h *ContentHandler) GetComments(c *gin.Context) {
	postID := c.Param("postID")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	comments, err := h.contentService.GetComments(c.Request.Context(), postID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// RecordView is the handler for recording a view on a post.
func (h *ContentHandler) RecordView(c *gin.Context) {
	postID := c.Param("postID")
	if err := h.contentService.RecordView(c.Request.Context(), postID); err != nil {
		// We can choose to ignore errors here or log them, but we won't send an error response
		// to the client to avoid impacting user experience for a non-critical operation.
	}
	c.Status(http.StatusNoContent)
}
