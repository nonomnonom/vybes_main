package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Follow represents a follow relationship between two users.
type Follow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FollowerID  primitive.ObjectID `bson:"followerId" json:"followerId"`   // The user who is doing the following
	FollowingID primitive.ObjectID `bson:"followingId" json:"followingId"` // The user who is being followed
}