package middleware

import (
	"net/http"
	"strings"
	"vybes/internal/service"
	"vybes/pkg/cache"

	"github.com/gin-gonic/gin"
)

// WalletAuthMiddleware creates a Gin middleware for wallet session authentication.
func WalletAuthMiddleware(walletSecurityService service.WalletSecurityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("X-Wallet-Session")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wallet session header is missing"})
			return
		}

		sessionToken := strings.TrimSpace(authHeader)
		if sessionToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid wallet session token"})
			return
		}

		// Validate session token
		session, err := walletSecurityService.ValidateWalletSession(c.Request.Context(), sessionToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set session info in context for the handler to use
		c.Set("walletSession", session)
		c.Set("userID", session.UserID.Hex())
		c.Next()
	}
}

// WalletRateLimitMiddleware creates a Gin middleware for rate limiting wallet operations.
func WalletRateLimitMiddleware(cache cache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		// Rate limit: max 10 wallet operations per minute per user
		key := "wallet_rate_limit:" + userID.(string)
		count, err := cache.Incr(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			return
		}

		// Set expiry for the first request
		if count == 1 {
			cache.Expire(c.Request.Context(), key, 60) // 60 seconds
		}

		if count > 10 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Wallet operation rate limit exceeded. Try again later."})
			return
		}

		c.Next()
	}
}

// WalletAuditMiddleware creates a Gin middleware for logging wallet operations.
func WalletAuditMiddleware(walletSecurityService service.WalletSecurityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client info
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// Store in context for handlers to use
		c.Set("clientIP", ipAddress)
		c.Set("userAgent", userAgent)

		c.Next()
	}
}