package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionType defines the type of reaction.
type ReactionType string

const (
	ReactionTypeLike ReactionType = "like"
)

// Reaction represents a reaction to a post.
type Reaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	PostID    primitive.ObjectID `bson:"postId" json:"postId"`
	Type      ReactionType       `bson:"type" json:"type"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
