package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the database.
type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	VID                 int64              `bson:"vid" json:"vid"`
	Name                string             `bson:"name" json:"name"`
	Email               string             `bson:"email" json:"email"`
	Username            string             `bson:"username,omitempty" json:"username,omitempty"`
	Password            string             `bson:"password" json:"-"`
	PFPURL              string             `bson:"pfpUrl,omitempty" json:"pfpUrl,omitempty"`
	BannerURL           string             `bson:"bannerUrl,omitempty" json:"bannerUrl,omitempty"`
	Bio                 string             `bson:"bio,omitempty" json:"bio,omitempty"`
	WalletAddress       string             `bson:"walletAddress" json:"walletAddress"`
	EncryptedPrivateKey string             `bson:"encryptedPrivateKey" json:"-"`
	TotalLikeCount      int64              `bson:"totalLikeCount" json:"totalLikeCount"`
	PostCount           int64              `bson:"postCount" json:"postCount"`
	OTP                 string             `bson:"otp,omitempty" json:"-"`
	OTPExpires          time.Time          `bson:"otpExpires,omitempty" json:"-"`
}
