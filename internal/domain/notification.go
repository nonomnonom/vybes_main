package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType defines the type of notification.
type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeComment NotificationType = "comment"
	NotificationTypeFollow  NotificationType = "follow"
)

// Notification represents a user notification.
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`       // The user who receives the notification
	ActorID   primitive.ObjectID `bson:"actorId" json:"actorId"`     // The user who triggered the notification
	Type      NotificationType   `bson:"type" json:"type"`
	PostID    *primitive.ObjectID `bson:"postId,omitempty" json:"postId,omitempty"` // Optional, for like/comment
	Read      bool               `bson:"read" json:"read"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}