package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the database.
type User struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	VID                  int64              `bson:"vid" json:"vid"`
	Name                 string             `bson:"name" json:"name"`
	Email                string             `bson:"email" json:"email"`
	Username             string             `bson:"username,omitempty" json:"username,omitempty"`
	Password             string             `bson:"password" json:"-"`
	PFPURL               string             `bson:"pfpUrl,omitempty" json:"pfpUrl,omitempty"`
	BannerURL            string             `bson:"bannerUrl,omitempty" json:"bannerUrl,omitempty"`
	Bio                  string             `bson:"bio,omitempty" json:"bio,omitempty"`
	WalletAddress        string             `bson:"walletAddress" json:"walletAddress"`
	EncryptedPrivateKey  string             `bson:"encryptedPrivateKey" json:"-"`
	TotalLikeCount       int64              `bson:"totalLikeCount" json:"totalLikeCount"`
	PostCount            int64              `bson:"postCount" json:"postCount"`
	OTP                  string             `bson:"otp,omitempty" json:"-"`
	OTPExpires           time.Time          `bson:"otpExpires,omitempty" json:"-"`
	// Enhanced security fields
	WalletNonce          uint64             `bson:"walletNonce" json:"-"`
	LastWalletAccess     time.Time          `bson:"lastWalletAccess,omitempty" json:"-"`
	FailedLoginAttempts  int                `bson:"failedLoginAttempts" json:"-"`
	AccountLockedUntil   time.Time          `bson:"accountLockedUntil,omitempty" json:"-"`
}

// WalletSession represents a temporary wallet access session
type WalletSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	SessionToken string             `bson:"sessionToken" json:"sessionToken"`
	ExpiresAt    time.Time          `bson:"expiresAt" json:"expiresAt"`
	IPAddress    string             `bson:"ipAddress" json:"ipAddress"`
	UserAgent    string             `bson:"userAgent" json:"userAgent"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	LastUsedAt   time.Time          `bson:"lastUsedAt" json:"lastUsedAt"`
}

// WalletAuditLog represents audit trail for wallet operations
type WalletAuditLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	Action      string             `bson:"action" json:"action"`
	TxHash      string             `bson:"txHash,omitempty" json:"txHash,omitempty"`
	Nonce       uint64             `bson:"nonce,omitempty" json:"nonce,omitempty"`
	Amount      string             `bson:"amount,omitempty" json:"amount,omitempty"`
	ToAddress   string             `bson:"toAddress,omitempty" json:"toAddress,omitempty"`
	IPAddress   string             `bson:"ipAddress" json:"ipAddress"`
	UserAgent   string             `bson:"userAgent" json:"userAgent"`
	Status      string             `bson:"status" json:"status"`
	ErrorMsg    string             `bson:"errorMsg,omitempty" json:"errorMsg,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}