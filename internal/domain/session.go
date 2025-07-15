package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	ClientIP     string             `bson:"client_ip" json:"client_ip"`
	IsBlocked    bool               `bson:"is_blocked" json:"is_blocked"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}