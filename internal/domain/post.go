package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostVisibility defines the visibility level of a post.
type PostVisibility string

const (
	VisibilityPublic  PostVisibility = "public"
	VisibilityFriends PostVisibility = "friends" // Friends are followers
	VisibilityPrivate PostVisibility = "private"
)

// Post represents a user-generated post, which is a container for content.
type Post struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         primitive.ObjectID  `bson:"userId" json:"userId"`
	ContentID      primitive.ObjectID  `bson:"contentId" json:"contentId"`
	OriginalPostID *primitive.ObjectID `bson:"originalPostId,omitempty" json:"originalPostId,omitempty"` // Pointer to distinguish null from empty
	LikeCount      int64               `bson:"likeCount" json:"likeCount"`
	CommentCount   int64               `bson:"commentCount" json:"commentCount"`
	RepostCount    int64               `bson:"repostCount" json:"repostCount"`
	ViewCount      int64               `bson:"viewCount" json:"viewCount"`
	Visibility     PostVisibility      `bson:"visibility" json:"visibility"`
	CreatedAt      time.Time           `bson:"createdAt" json:"createdAt"`
}