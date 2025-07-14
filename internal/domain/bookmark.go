package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Bookmark represents a user's bookmark on a post.
type Bookmark struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	PostID    primitive.ObjectID `bson:"postId" json:"postId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}