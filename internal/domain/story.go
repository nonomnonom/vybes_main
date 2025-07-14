package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Story represents a user's story.
type Story struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	MediaURL  string             `bson:"mediaUrl" json:"mediaUrl"`
	MediaType string             `bson:"mediaType" json:"mediaType"` // e.g., "image/jpeg", "video/mp4"
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}