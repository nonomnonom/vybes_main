# Tipping Service API Documentation

## Overview

The Tipping Service allows users to send tips to each other using $VYB tokens. Each user has a weekly allowance of 10,000 $VYB that resets automatically every Monday.

## Features

- **Weekly Allowance**: 10,000 $VYB per user per week
- **Direct Tipping**: Send tips directly to other users
- **Content Tipping**: Tip content creators via comments using `$amount$vyb` format
- **Automatic Reset**: Allowances reset every Monday at 00:00
- **Statistics Tracking**: Track total and weekly tipping statistics

## API Endpoints

### Authentication

All tipping endpoints require authentication. Include the JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### 1. Send Tip

**POST** `/api/v1/tipping/send`

Send a tip from the authenticated user to another user.

**Request Body:**
```json
{
  "toUserId": "507f1f77bcf86cd799439011",
  "amount": 100,
  "message": "Great content!"
}
```

**Response:**
```json
{
  "message": "Tip sent successfully",
  "tip": {
    "id": "507f1f77bcf86cd799439012",
    "fromUserId": "507f1f77bcf86cd799439013",
    "toUserId": "507f1f77bcf86cd799439011",
    "amount": 100,
    "message": "Great content!",
    "status": "COMPLETED",
    "createdAt": "2024-01-15T10:30:00Z",
    "completedAt": "2024-01-15T10:30:00Z"
  }
}
```

### 2. Get Allowance

**GET** `/api/v1/tipping/allowance`

Get the current user's weekly tipping allowance.

**Response:**
```json
{
  "userId": "507f1f77bcf86cd799439013",
  "weeklyLimit": 10000,
  "usedAmount": 2500,
  "remaining": 7500,
  "weekStart": "2024-01-15T00:00:00Z",
  "lastReset": "2024-01-15T00:00:00Z"
}
```

### 3. Get User Tips

**GET** `/api/v1/tipping/history?limit=50`

Get the authenticated user's tipping history (sent and received).

**Query Parameters:**
- `limit` (optional): Number of tips to return (default: 50, max: 100)

**Response:**
```json
{
  "tips": [
    {
      "id": "507f1f77bcf86cd799439012",
      "fromUserId": "507f1f77bcf86cd799439013",
      "toUserId": "507f1f77bcf86cd799439011",
      "amount": 100,
      "message": "Great content!",
      "status": "COMPLETED",
      "contentId": "507f1f77bcf86cd799439014",
      "commentId": "507f1f77bcf86cd799439015",
      "createdAt": "2024-01-15T10:30:00Z",
      "completedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

### 4. Get Content Tips

**GET** `/api/v1/tipping/content/:contentId`

Get all tips for a specific content.

**Response:**
```json
{
  "tips": [
    {
      "id": "507f1f77bcf86cd799439012",
      "fromUserId": "507f1f77bcf86cd799439013",
      "toUserId": "507f1f77bcf86cd799439011",
      "amount": 100,
      "message": "Great content!",
      "status": "COMPLETED",
      "contentId": "507f1f77bcf86cd799439014",
      "commentId": "507f1f77bcf86cd799439015",
      "createdAt": "2024-01-15T10:30:00Z",
      "completedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

### 5. Get Tip Statistics

**GET** `/api/v1/tipping/stats`

Get tipping statistics for the authenticated user.

**Response:**
```json
{
  "userId": "507f1f77bcf86cd799439013",
  "totalReceived": 5000,
  "totalSent": 2500,
  "weeklyReceived": 1000,
  "weeklySent": 500,
  "lastUpdated": "2024-01-15T10:30:00Z"
}
```

## Content Tipping via Comments

Users can tip content creators by including a tip pattern in their comments:

### Tip Pattern Format

```
$amount$vyb
```

**Examples:**
- `$100$vyb` - Tip 100 VYB
- `$500$vyb` - Tip 500 VYB
- `Great post! $250$vyb` - Comment with tip

### How It Works

1. User comments on content with tip pattern
2. System automatically detects the pattern
3. Tip is processed from commenter to content creator
4. Tip is linked to both the content and comment

### Example Comment

```
This is amazing content! $100$vyb
```

This will automatically send 100 VYB from the commenter to the content creator.

## Error Responses

### Insufficient Allowance
```json
{
  "error": "insufficient weekly allowance. Used: 9500, Limit: 10000, Requested: 1000"
}
```

### Invalid Amount
```json
{
  "error": "tip amount must be positive"
}
```

### User Not Found
```json
{
  "error": "recipient user not found"
}
```

### Invalid Tip Pattern
```json
{
  "error": "no valid tip pattern found in text"
}
```

## Weekly Reset Schedule

- **Allowance Reset**: Every Monday at 00:00 UTC
- **Stats Reset**: Every Monday at 00:01 UTC
- **Reset Process**: Automatic via cron jobs

## Database Collections

### tipping_allowances
Stores weekly allowance information for each user.

### tips
Stores all tip transactions.

### tip_stats
Stores tipping statistics for each user.

## Rate Limits

- Maximum tip amount: 10,000 VYB per transaction
- Weekly allowance: 10,000 VYB per user
- No limit on number of tips per week (within allowance)

## Security Considerations

- All endpoints require authentication
- Tips are processed atomically
- Allowance validation prevents overspending
- Comment tips are processed asynchronously to avoid blocking comment creation