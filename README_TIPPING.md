# Vybes Tipping Service

A comprehensive tipping system for the Vybes platform that allows users to send tips using $VYB tokens.

## 🚀 Features

- **Weekly Allowance**: 10,000 $VYB per user per week (auto-reset every Monday)
- **Direct Tipping**: Send tips directly to other users
- **Content Tipping**: Tip content creators via comments using `$amount$vyb` format
- **Automatic Processing**: Comment tips are processed automatically
- **Statistics Tracking**: Comprehensive tipping statistics and history
- **Cron Jobs**: Automatic weekly allowance and stats reset

## 📋 Requirements

- Go 1.23.0+
- MongoDB
- Existing Vybes application structure

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Layer     │    │  Service Layer  │    │ Repository Layer│
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Handler   │ │───▶│   Service    │ │───▶│ Repository   │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Middleware    │    │    MongoDB      │
                       │                 │    │                 │
                       │ Comment Tip     │    │ Collections:    │
                       │ Processing      │    │ - tips          │
                       └─────────────────┘    │ - allowances    │
                                              │ - tip_stats     │
                                              └─────────────────┘
```

## 📁 File Structure

```
internal/
├── domain/
│   └── tipping.go              # Domain models
├── repository/
│   └── tipping_repository.go   # Database operations
├── service/
│   └── tipping_service.go      # Business logic
├── handler/
│   ├── tipping_handler.go      # HTTP handlers
│   └── tipping_routes.go       # Route definitions
└── middleware/
    └── comment_tip_middleware.go # Comment tip processing

docs/
├── TIPPING_API.md              # API documentation
└── TIPPING_INTEGRATION.md      # Integration guide

test/
└── tipping_test.go             # Unit tests

examples/
└── tipping_example.go          # Usage examples
```

## 🚀 Quick Start

### 1. Install Dependencies

```bash
go get github.com/stretchr/testify
```

### 2. Integration

Add the tipping service to your main application:

```go
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

// Setup routes
api := router.Group("/api/v1")
handler.SetupTippingRoutes(api, tippingHandler, authMiddleware)
```

### 3. Run Tests

```bash
go test ./test/tipping_test.go -v
```

## 📚 API Reference

### Authentication

All endpoints require JWT authentication:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tipping/send` | Send a tip to another user |
| GET | `/api/v1/tipping/allowance` | Get current user's allowance |
| GET | `/api/v1/tipping/history` | Get user's tipping history |
| GET | `/api/v1/tipping/stats` | Get tipping statistics |
| GET | `/api/v1/tipping/content/:id` | Get tips for specific content |

### Example Usage

#### Send a Tip

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

#### Comment with Tip

```bash
curl -X POST http://localhost:8080/api/v1/comments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "postId": "507f1f77bcf86cd799439014",
    "text": "Amazing content! $100$vyb"
  }'
```

## 💡 Tip Patterns

Users can tip via comments using the following pattern:

```
$amount$vyb
```

**Examples:**
- `$100$vyb` - Tip 100 VYB
- `$500$vyb` - Tip 500 VYB
- `Great post! $250$vyb` - Comment with tip

## 🔄 Weekly Reset

The system automatically resets allowances and stats every Monday:
- **Allowance Reset**: Monday 00:00 UTC
- **Stats Reset**: Monday 00:01 UTC

## 🗄️ Database Collections

### tipping_allowances
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "weeklyLimit": 10000,
  "usedAmount": 2500,
  "weekStart": ISODate,
  "lastReset": ISODate,
  "createdAt": ISODate,
  "updatedAt": ISODate
}
```

### tips
```json
{
  "_id": ObjectId,
  "fromUserId": ObjectId,
  "toUserId": ObjectId,
  "amount": 100,
  "contentId": ObjectId,
  "commentId": ObjectId,
  "message": "Great content!",
  "status": "COMPLETED",
  "createdAt": ISODate,
  "completedAt": ISODate
}
```

### tip_stats
```json
{
  "userId": ObjectId,
  "totalReceived": 5000,
  "totalSent": 2500,
  "weeklyReceived": 1000,
  "weeklySent": 500,
  "lastUpdated": ISODate
}
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./test/ -v

# Run specific test
go test ./test/tipping_test.go -v

# Run with coverage
go test ./test/ -v -cover
```

## 🔧 Configuration

### Environment Variables

No additional environment variables required. Uses existing database connections.

### Cron Jobs

The service includes automatic cron jobs for:
- Weekly allowance reset (Monday 00:00 UTC)
- Weekly stats reset (Monday 00:01 UTC)

## 🛡️ Security

- **Authentication**: All endpoints require valid JWT tokens
- **Validation**: All amounts validated as positive integers
- **Allowance Enforcement**: Weekly limits strictly enforced
- **Atomic Operations**: Tips processed atomically
- **Rate Limiting**: Consider adding rate limits for production

## 📊 Monitoring

### Key Metrics

- Number of tips per day/week
- Total VYB tipped
- Users with highest tips received
- Failed tip transactions
- Allowance utilization

### Logs

The service logs important events:
- Tip transactions
- Allowance resets
- Weekly stats resets
- Processing errors

## 🚨 Troubleshooting

### Common Issues

1. **"insufficient weekly allowance" error**
   - Check user's remaining allowance
   - Verify weekly reset schedule

2. **Comment tips not processing**
   - Verify `$amount$vyb` pattern in comment
   - Check comment creation success
   - Review processing logs

3. **Cron jobs not running**
   - Verify cron service is started
   - Check server timezone
   - Review cron logs

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is part of the Vybes platform.

## 🆘 Support

For issues or questions:
1. Check the logs for error messages
2. Review the API documentation
3. Run the test suite
4. Contact the development team

---

**Happy Tipping! 🎉**