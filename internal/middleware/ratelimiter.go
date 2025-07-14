package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterMiddleware creates a new rate limiter middleware.
func RateLimiterMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*rate.Limiter)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if !v.Allow() {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		limiter, exists := visitors[ip]
		if !exists {
			limiter = rate.NewLimiter(r, b)
			visitors[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}