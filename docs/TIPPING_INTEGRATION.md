# Tipping Service Integration Guide

## Overview

This guide explains how to integrate the tipping service into your existing Vybes application.

## Prerequisites

- Go 1.23.0 or higher
- MongoDB database
- Existing Vybes application structure

## Integration Steps

### 1. Update Dependencies

Add testify for testing (if not already present):

```bash
go get github.com/stretchr/testify
```

### 2. Database Setup

The tipping service creates three new collections in MongoDB:

- `tipping_allowances` - Weekly allowance tracking
- `tips` - Tip transactions
- `tip_stats` - User tipping statistics

No additional setup required - collections are created automatically.

### 3. Service Integration

#### Update your main application file (e.g., `cmd/main.go`):

```go
package main

import (
    "vybes/internal/handler"
    "vybes/internal/middleware"
    "vybes/internal/repository"
    "vybes/internal/service"
    // ... other imports
)

func main() {
    // ... existing setup code ...
    
    // Initialize repositories
    tippingRepo := repository.NewTippingRepository(db)
    
    // Initialize services
    tippingService := service.NewTippingService(
        tippingRepo,
        userRepo,
        commentRepo,
        contentRepo,
    )
    
    // Initialize handlers
    tippingHandler := handler.NewTippingHandler(tippingService)
    
    // Initialize middleware
    commentTipMiddleware := middleware.NewCommentTipMiddleware(tippingService)
    
    // Update cron service
    cronService := service.NewCronService(cfg, storyRepo, tippingService, storage)
    
    // ... rest of setup ...
    
    // Setup routes
    api := router.Group("/api/v1")
    handler.SetupTippingRoutes(api, tippingHandler, authMiddleware)
    handler.SetupInternalTippingRoutes(api, tippingHandler)
    
    // ... rest of main function ...
}
```

### 4. Comment Integration

To enable tipping via comments, update your comment creation endpoint:

```go
// In your comment handler
func (h *CommentHandler) CreateComment(c *gin.Context) {
    // ... existing comment creation logic ...
    
    // After successful comment creation, check for tips
    if hasTip, exists := c.Get("hasTip"); exists && hasTip.(bool) {
        // Process tip asynchronously
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            
            err := h.tippingService.ProcessCommentTip(ctx, comment.ID)
            if err != nil {
                // Log error but don't fail the request
                log.Error().Err(err).Msg("Failed to process comment tip")
            }
        }()
    }
    
    // ... return response ...
}
```

### 5. Add Middleware to Comment Routes

```go
// In your route setup
commentTipMiddleware := middleware.NewCommentTipMiddleware(tippingService)

comments := api.Group("/comments")
comments.Use(authMiddleware)
{
    comments.POST("/", commentTipMiddleware.ValidateCommentTip(), commentHandler.CreateComment)
    // ... other comment routes ...
}
```

## Configuration

### Environment Variables

No additional environment variables are required. The service uses existing database connections.

### Weekly Reset Schedule

The cron jobs automatically reset allowances and stats every Monday:
- Allowance reset: Monday 00:00 UTC
- Stats reset: Monday 00:01 UTC

## API Usage Examples

### Send a Direct Tip

```bash
curl -X POST http://localhost:8080/api/v1/tipping/send \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "toUserId": "507f1f77bcf86cd799439011",
    "amount": 100,
    "message": "Great content!"
  }'
```

### Get User's Allowance

```bash
curl -X GET http://localhost:8080/api/v1/tipping/allowance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Get Tipping History

```bash
curl -X GET "http://localhost:8080/api/v1/tipping/history?limit=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Comment with Tip

```bash
curl -X POST http://localhost:8080/api/v1/comments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "postId": "507f1f77bcf86cd799439014",
    "text": "Amazing content! $100$vyb"
  }'
```

## Testing

Run the tipping service tests:

```bash
go test ./test/tipping_test.go -v
```

## Monitoring

### Logs

The service logs important events:
- Tip transactions
- Allowance resets
- Weekly stats resets
- Errors during tip processing

### Metrics to Monitor

- Number of tips per day/week
- Total VYB tipped
- Users with highest tips received
- Failed tip transactions

## Security Considerations

1. **Authentication**: All tipping endpoints require valid JWT tokens
2. **Rate Limiting**: Consider adding rate limits to prevent abuse
3. **Validation**: All amounts are validated to be positive integers
4. **Allowance Enforcement**: Weekly limits are strictly enforced
5. **Atomic Operations**: Tips are processed atomically to prevent double-spending

## Troubleshooting

### Common Issues

1. **"insufficient weekly allowance" error**
   - Check if user has enough allowance remaining
   - Verify weekly reset schedule

2. **Comment tips not processing**
   - Check if comment contains valid `$amount$vyb` pattern
   - Verify comment creation was successful
   - Check logs for processing errors

3. **Cron jobs not running**
   - Verify cron service is started
   - Check server timezone settings
   - Review cron logs

### Debug Mode

Enable debug logging by setting log level to debug in your configuration.

## Migration from Existing System

If you have an existing tipping system:

1. **Data Migration**: Export existing tip data and import to new collections
2. **Gradual Rollout**: Deploy to staging first, then production
3. **Backup**: Always backup existing data before migration
4. **Monitoring**: Monitor closely during initial deployment

## Support

For issues or questions:
1. Check the logs for error messages
2. Review the API documentation
3. Run the test suite to verify functionality
4. Contact the development team