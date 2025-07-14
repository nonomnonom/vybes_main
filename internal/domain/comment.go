package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a comment on a post.
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	PostID    primitive.ObjectID `bson:"postId" json:"postId"`
	Text      string             `bson:"text" json:"text"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}