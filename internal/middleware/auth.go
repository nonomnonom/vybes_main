package middleware

import (
	"net/http"
	"strings"
	"vybes/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
// This middleware validates JWT tokens from the Authorization header
// and extracts user information for use in subsequent handlers.
// The middleware supports both "Bearer" and "Token" authorization schemes.
//
// Parameters:
//   - jwtSecret: The secret key used to sign and verify JWT tokens
//
// Returns:
//   - gin.HandlerFunc: A Gin middleware function that validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Parse authorization header to extract token
		// Supports both "Bearer <token>" and "Token <token>" formats
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "Token") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Warn().Err(err).Msg("JWT token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user ID from claims
			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}

			// Parse user ID string to ObjectID
			userID, err := domain.ParseObjectID(userIDStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			// Set user ID in context for the handler to use
			c.Set("user_id", userID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
