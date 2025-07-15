package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TippingAllowance represents the weekly allowance for a user
type TippingAllowance struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	WeeklyLimit  int64              `bson:"weeklyLimit" json:"weeklyLimit"`   // 10000 VYB
	UsedAmount   int64              `bson:"usedAmount" json:"usedAmount"`     // Amount used this week
	WeekStart    time.Time          `bson:"weekStart" json:"weekStart"`       // Start of current week
	LastReset    time.Time          `bson:"lastReset" json:"lastReset"`       // Last time allowance was reset
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Tip represents a tip transaction
type Tip struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FromUserID  primitive.ObjectID `bson:"fromUserId" json:"fromUserId"`
	ToUserID    primitive.ObjectID `bson:"toUserId" json:"toUserId"`
	Amount      int64              `bson:"amount" json:"amount"`           // Amount in VYB
	ContentID   *primitive.ObjectID `bson:"contentId,omitempty" json:"contentId,omitempty"` // Optional: if tipping via content
	CommentID   *primitive.ObjectID `bson:"commentId,omitempty" json:"commentId,omitempty"` // Optional: if tipping via comment
	Message     string             `bson:"message,omitempty" json:"message,omitempty"`     // Optional message
	Status      TipStatus          `bson:"status" json:"status"`           // PENDING, COMPLETED, FAILED
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	CompletedAt *time.Time         `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
}

// TipStatus represents the status of a tip transaction
type TipStatus string

const (
	TipStatusPending   TipStatus = "PENDING"
	TipStatusCompleted TipStatus = "COMPLETED"
	TipStatusFailed    TipStatus = "FAILED"
)

// TipStats represents tipping statistics for a user
type TipStats struct {
	UserID           primitive.ObjectID `bson:"userId" json:"userId"`
	TotalReceived    int64              `bson:"totalReceived" json:"totalReceived"`
	TotalSent        int64              `bson:"totalSent" json:"totalSent"`
	WeeklyReceived   int64              `bson:"weeklyReceived" json:"weeklyReceived"`
	WeeklySent       int64              `bson:"weeklySent" json:"weeklySent"`
	LastUpdated      time.Time          `bson:"lastUpdated" json:"lastUpdated"`
}