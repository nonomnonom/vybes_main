package handler

import (
	"vybes/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupTippingRoutes sets up all tipping-related routes
func SetupTippingRoutes(router *gin.RouterGroup, tippingHandler *TippingHandler, authMiddleware gin.HandlerFunc) {
	// All tipping routes require authentication
	tipping := router.Group("/tipping")
	tipping.Use(authMiddleware)
	
	{
		// Send a tip to another user
		tipping.POST("/send", tippingHandler.SendTip)
		
		// Get current user's allowance
		tipping.GET("/allowance", tippingHandler.GetAllowance)
		
		// Get user's tipping history
		tipping.GET("/history", tippingHandler.GetUserTips)
		
		// Get tipping statistics
		tipping.GET("/stats", tippingHandler.GetTipStats)
		
		// Get tips for a specific content
		tipping.GET("/content/:contentId", tippingHandler.GetContentTips)
	}
}

// SetupInternalTippingRoutes sets up internal tipping routes (for system use)
func SetupInternalTippingRoutes(router *gin.RouterGroup, tippingHandler *TippingHandler) {
	internal := router.Group("/internal/tipping")
	
	{
		// Process tip from comment (called by comment middleware)
		internal.POST("/process-comment/:commentId", tippingHandler.ProcessCommentTip)
	}
}