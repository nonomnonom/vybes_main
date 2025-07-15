package middleware

import (
	"context"
	"regexp"
	"time"

	"vybes/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentTipMiddleware struct {
	tippingService *service.TippingService
}

func NewCommentTipMiddleware(tippingService *service.TippingService) *CommentTipMiddleware {
	return &CommentTipMiddleware{
		tippingService: tippingService,
	}
}

// TipPattern matches $100$vyb format
var TipPattern = regexp.MustCompile(`\$(\d+)\$vyb`)

// ProcessCommentTip processes tips from comments asynchronously
func (m *CommentTipMiddleware) ProcessCommentTip() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue with the request first
		c.Next()
		
		// Only process if the request was successful
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Get the comment ID from the response or request
			commentIDStr := c.Param("commentId")
			if commentIDStr == "" {
				// Try to get from response body if it's a comment creation
				// This would need to be implemented based on your comment creation response
				return
			}
			
			commentID, err := primitive.ObjectIDFromHex(commentIDStr)
			if err != nil {
				return
			}
			
			// Process tip asynchronously
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				
				err := m.tippingService.ProcessCommentTip(ctx, commentID)
				if err != nil {
					// Log error but don't fail the request
					// You might want to add proper logging here
					return
				}
			}()
		}
	}
}

// ValidateCommentTip validates if a comment contains a valid tip pattern
func (m *CommentTipMiddleware) ValidateCommentTip() gin.HandlerFunc {
	return func(c *gin.Context) {
		var comment struct {
			Text string `json:"text" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.Next()
			return
		}
		
		// Check if comment contains tip pattern
		if TipPattern.MatchString(comment.Text) {
			// Add flag to context for later processing
			c.Set("hasTip", true)
		}
		
		c.Next()
	}
}