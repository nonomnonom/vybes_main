package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentType defines the type of content.
type ContentType string

const (
	ContentTypeVideo ContentType = "video"
	ContentTypeImage ContentType = "image"
)

// Content represents a piece of media content.
type Content struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Type      ContentType        `bson:"type" json:"type"`
	URL       string             `bson:"url" json:"url"`
	Caption   string             `bson:"caption,omitempty" json:"caption,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}