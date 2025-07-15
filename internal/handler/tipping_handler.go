package handler

import (
	"net/http"
	"strconv"

	"vybes/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TippingHandler struct {
	tippingService *service.TippingService
}

func NewTippingHandler(tippingService *service.TippingService) *TippingHandler {
	return &TippingHandler{
		tippingService: tippingService,
	}
}

// Request/Response structures
type SendTipRequest struct {
	ToUserID string `json:"toUserId" binding:"required"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Message  string `json:"message,omitempty"`
}

type TipResponse struct {
	ID          string  `json:"id"`
	FromUserID  string  `json:"fromUserId"`
	ToUserID    string  `json:"toUserId"`
	Amount      int64   `json:"amount"`
	Message     string  `json:"message,omitempty"`
	Status      string  `json:"status"`
	ContentID   *string `json:"contentId,omitempty"`
	CommentID   *string `json:"commentId,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt *string `json:"completedAt,omitempty"`
}

type AllowanceResponse struct {
	UserID      string `json:"userId"`
	WeeklyLimit int64  `json:"weeklyLimit"`
	UsedAmount  int64  `json:"usedAmount"`
	Remaining   int64  `json:"remaining"`
	WeekStart   string `json:"weekStart"`
	LastReset   string `json:"lastReset"`
}

type TipStatsResponse struct {
	UserID         string `json:"userId"`
	TotalReceived  int64  `json:"totalReceived"`
	TotalSent      int64  `json:"totalSent"`
	WeeklyReceived int64  `json:"weeklyReceived"`
	WeeklySent     int64  `json:"weeklySent"`
	LastUpdated    string `json:"lastUpdated"`
}

// SendTip sends a tip from the authenticated user to another user
func (h *TippingHandler) SendTip(c *gin.Context) {
	var req SendTipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get authenticated user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	fromUserID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	toUserID, err := primitive.ObjectIDFromHex(req.ToUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipient user ID"})
		return
	}

	// Send tip
	tip, err := h.tippingService.SendTip(c.Request.Context(), fromUserID, toUserID, req.Amount, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	response := TipResponse{
		ID:         tip.ID.Hex(),
		FromUserID: tip.FromUserID.Hex(),
		ToUserID:   tip.ToUserID.Hex(),
		Amount:     tip.Amount,
		Message:    tip.Message,
		Status:     string(tip.Status),
		CreatedAt:  tip.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if tip.CompletedAt != nil {
		completedAt := tip.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	if tip.ContentID != nil {
		contentID := tip.ContentID.Hex()
		response.ContentID = &contentID
	}

	if tip.CommentID != nil {
		commentID := tip.CommentID.Hex()
		response.CommentID = &commentID
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tip sent successfully",
		"tip":     response,
	})
}

// GetAllowance gets the current user's tipping allowance
func (h *TippingHandler) GetAllowance(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get allowance
	allowance, err := h.tippingService.GetOrCreateAllowance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := AllowanceResponse{
		UserID:      allowance.UserID.Hex(),
		WeeklyLimit: allowance.WeeklyLimit,
		UsedAmount:  allowance.UsedAmount,
		Remaining:   allowance.WeeklyLimit - allowance.UsedAmount,
		WeekStart:   allowance.WeekStart.Format("2006-01-02T15:04:05Z"),
		LastReset:   allowance.LastReset.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserTips gets all tips for the authenticated user
func (h *TippingHandler) GetUserTips(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = 50
	}

	// Get tips
	tips, err := h.tippingService.GetUserTips(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]TipResponse, len(tips))
	for i, tip := range tips {
		response := TipResponse{
			ID:         tip.ID.Hex(),
			FromUserID: tip.FromUserID.Hex(),
			ToUserID:   tip.ToUserID.Hex(),
			Amount:     tip.Amount,
			Message:    tip.Message,
			Status:     string(tip.Status),
			CreatedAt:  tip.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if tip.CompletedAt != nil {
			completedAt := tip.CompletedAt.Format("2006-01-02T15:04:05Z")
			response.CompletedAt = &completedAt
		}

		if tip.ContentID != nil {
			contentID := tip.ContentID.Hex()
			response.ContentID = &contentID
		}

		if tip.CommentID != nil {
			commentID := tip.CommentID.Hex()
			response.CommentID = &commentID
		}

		responses[i] = response
	}

	c.JSON(http.StatusOK, gin.H{
		"tips": responses,
		"count": len(responses),
	})
}

// GetContentTips gets all tips for a specific content
func (h *TippingHandler) GetContentTips(c *gin.Context) {
	contentIDStr := c.Param("contentId")
	contentID, err := primitive.ObjectIDFromHex(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content ID"})
		return
	}

	// Get tips
	tips, err := h.tippingService.GetContentTips(c.Request.Context(), contentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]TipResponse, len(tips))
	for i, tip := range tips {
		response := TipResponse{
			ID:         tip.ID.Hex(),
			FromUserID: tip.FromUserID.Hex(),
			ToUserID:   tip.ToUserID.Hex(),
			Amount:     tip.Amount,
			Message:    tip.Message,
			Status:     string(tip.Status),
			CreatedAt:  tip.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if tip.CompletedAt != nil {
			completedAt := tip.CompletedAt.Format("2006-01-02T15:04:05Z")
			response.CompletedAt = &completedAt
		}

		if tip.ContentID != nil {
			contentID := tip.ContentID.Hex()
			response.ContentID = &contentID
		}

		if tip.CommentID != nil {
			commentID := tip.CommentID.Hex()
			response.CommentID = &commentID
		}

		responses[i] = response
	}

	c.JSON(http.StatusOK, gin.H{
		"tips": responses,
		"count": len(responses),
	})
}

// GetTipStats gets tipping statistics for the authenticated user
func (h *TippingHandler) GetTipStats(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get stats
	stats, err := h.tippingService.GetTipStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := TipStatsResponse{
		UserID:         stats.UserID.Hex(),
		TotalReceived:  stats.TotalReceived,
		TotalSent:      stats.TotalSent,
		WeeklyReceived: stats.WeeklyReceived,
		WeeklySent:     stats.WeeklySent,
		LastUpdated:    stats.LastUpdated.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// ProcessCommentTip processes a tip from a comment (internal endpoint)
func (h *TippingHandler) ProcessCommentTip(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	// Process the tip
	err = h.tippingService.ProcessCommentTip(c.Request.Context(), commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment tip processed successfully",
	})
}