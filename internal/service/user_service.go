package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/cache"
	"vybes/pkg/evm"
	"vybes/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ... (structs are unchanged)
type UpdateProfilePayload struct {
	Username  *string `json:"username"`
	PFPURL    *string `json:"pfpUrl"`
	BannerURL *string `json:"bannerUrl"`
	Bio       *string `json:"bio"`
}
type UserProfileResponse struct {
	domain.User
	FollowerCount   int64  `json:"followerCount"`
	FollowingCount  int64  `json:"followingCount"`
	FriendshipStatus string `json:"friendshipStatus"`
}

// UserService defines the interface for user business logic.
type UserService interface {
	Register(ctx context.Context, name, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	GetUserProfile(ctx context.Context, viewerID, username string) (*UserProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, payload UpdateProfilePayload) (*domain.User, error)
	ExportPrivateKey(ctx context.Context, userID, password string) (string, error)
	PersonalSign(ctx context.Context, userID, password, message string) (string, error)
	SignTransaction(ctx context.Context, userID, password string, tx *types.Transaction) (*types.Transaction, error)
	SendTransaction(ctx context.Context, userID, password string, tx *types.Transaction) (common.Hash, error)
	SignTypedDataV4(ctx context.Context, userID, password string, typedData apitypes.TypedData) (string, error)
	Secp256k1Sign(ctx context.Context, userID, password, hash string) (string, error)
	RequestOTP(ctx context.Context, email string) error
	VerifyOTPAndResetPassword(ctx context.Context, email, otp, newPassword string) error
}

type userService struct {
	userRepo            repository.UserRepository
	followRepo          repository.FollowRepository
	counterRepo         repository.CounterRepository
	walletService       WalletService
	emailService        EmailService
	cache               cache.Client
	jwtSecret           string
	walletEncryptionKey string
}

// NewUserService creates a new user service.
func NewUserService(userRepo repository.UserRepository, followRepo repository.FollowRepository, counterRepo repository.CounterRepository, walletService WalletService, emailService EmailService, cache cache.Client, jwtSecret, walletEncryptionKey string) UserService {
	return &userService{
		userRepo:            userRepo,
		followRepo:          followRepo,
		counterRepo:         counterRepo,
		walletService:       walletService,
		emailService:        emailService,
		cache:               cache,
		jwtSecret:           jwtSecret,
		walletEncryptionKey: walletEncryptionKey,
	}
}

// ... (Register, Login are unchanged)
func (s *userService) Register(ctx context.Context, name, email, password string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	nextVID, err := s.counterRepo.GetNextSequence(ctx, "user_vid")
	if err != nil {
		return errors.New("could not generate user VID")
	}
	walletAddress, encryptedPrivateKey, err := s.walletService.CreateWallet()
	if err != nil {
		return errors.New("could not create user wallet")
	}
	user := &domain.User{
		ID:                  primitive.NewObjectID(),
		VID:                 nextVID,
		Name:                name,
		Email:               email,
		Password:            hashedPassword,
		WalletAddress:       walletAddress,
		EncryptedPrivateKey: encryptedPrivateKey,
	}
	return s.userRepo.Create(ctx, user)
}
func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}
	claims := jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUserProfile retrieves a user's public profile and friendship status, with caching.
func (s *userService) GetUserProfile(ctx context.Context, viewerIDStr, username string) (*UserProfileResponse, error) {
	cacheKey := fmt.Sprintf("profile:%s:viewer:%s", username, viewerIDStr)

	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var profile UserProfileResponse
		if json.Unmarshal([]byte(cachedData), &profile) == nil {
			return &profile, nil
		}
	}
	if err != redis.Nil {
		// Log non-nil errors from Redis
	}

	viewerID, err := primitive.ObjectIDFromHex(viewerIDStr)
	if err != nil {
		return nil, errors.New("invalid viewer ID format")
	}

	profileUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if profileUser == nil {
		return nil, errors.New("user not found")
	}

	followerCount, err := s.followRepo.GetFollowerCount(ctx, profileUser.ID)
	if err != nil {
		return nil, err
	}

	followingCount, err := s.followRepo.GetFollowingCount(ctx, profileUser.ID)
	if err != nil {
		return nil, err
	}

	viewerIsFollowing, err := s.followRepo.IsFollowing(ctx, viewerID, profileUser.ID)
	if err != nil {
		return nil, err
	}

	profileUserIsFollowing, err := s.followRepo.IsFollowing(ctx, profileUser.ID, viewerID)
	if err != nil {
		return nil, err
	}

	var friendshipStatus string
	if viewerIsFollowing && profileUserIsFollowing {
		friendshipStatus = "mutual"
	} else if viewerIsFollowing {
		friendshipStatus = "following"
	} else if profileUserIsFollowing {
		friendshipStatus = "follows_you"
	} else {
		friendshipStatus = "none"
	}

	profileResponse := &UserProfileResponse{
		User:             *profileUser,
		FollowerCount:    followerCount,
		FollowingCount:   followingCount,
		FriendshipStatus: friendshipStatus,
	}

	jsonData, err := json.Marshal(profileResponse)
	if err == nil {
		s.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
	}

	return profileResponse, nil
}

// UpdateProfile updates a user's metadata and invalidates the cache.
func (s *userService) UpdateProfile(ctx context.Context, userID string, payload UpdateProfilePayload) (*domain.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	cacheKey := fmt.Sprintf("profile:%s:viewer:%s", user.Username, userID)
	s.cache.Del(ctx, cacheKey)


	if payload.Username != nil {
		user.Username = *payload.Username
	}
	if payload.PFPURL != nil {
		user.PFPURL = *payload.PFPURL
	}
	if payload.BannerURL != nil {
		user.BannerURL = *payload.BannerURL
	}
	if payload.Bio != nil {
		user.Bio = *payload.Bio
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ... (rest of the methods are unchanged)
func (s *userService) ExportPrivateKey(ctx context.Context, userIDStr, password string) (string, error) {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return "", errors.New("invalid user ID format")
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid password")
	}
	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return "", errors.New("could not decrypt private key")
	}
	return decryptedKey, nil
}
func (s *userService) PersonalSign(ctx context.Context, userIDStr, password, message string) (string, error) {
	decryptedKey, err := s.ExportPrivateKey(ctx, userIDStr, password)
	if err != nil {
		return "", err
	}
	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}
	return signer.PersonalSign([]byte(message))
}
func (s *userService) SignTransaction(ctx context.Context, userIDStr, password string, tx *types.Transaction) (*types.Transaction, error) {
	decryptedKey, err := s.ExportPrivateKey(ctx, userIDStr, password)
	if err != nil {
		return nil, err
	}
	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return nil, err
	}
	return signer.SignTransaction(tx)
}
func (s *userService) SignTypedDataV4(ctx context.Context, userIDStr, password string, typedData apitypes.TypedData) (string, error) {
	decryptedKey, err := s.ExportPrivateKey(ctx, userIDStr, password)
	if err != nil {
		return "", err
	}
	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}
	return signer.SignTypedDataV4(typedData)
}
func (s *userService) SendTransaction(ctx context.Context, userIDStr, password string, tx *types.Transaction) (common.Hash, error) {
	decryptedKey, err := s.ExportPrivateKey(ctx, userIDStr, password)
	if err != nil {
		return common.Hash{}, err
	}
	return s.walletService.SendTransaction(ctx, tx, decryptedKey)
}
func (s *userService) Secp256k1Sign(ctx context.Context, userIDStr, password, hashStr string) (string, error) {
	decryptedKey, err := s.ExportPrivateKey(ctx, userIDStr, password)
	if err != nil {
		return "", err
	}
	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}
	hash, err := hexutil.Decode(hashStr)
	if err != nil {
		return "", errors.New("invalid hash format")
	}
	return signer.Secp256k1Sign(hash)
}
func (s *userService) RequestOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	otp, err := utils.GenerateOTP(6)
	if err != nil {
		return errors.New("could not generate OTP")
	}
	user.OTP = otp
	user.OTPExpires = time.Now().Add(time.Minute * 5)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}
	return s.emailService.SendOTPEmail(user.Email, otp)
}
func (s *userService) VerifyOTPAndResetPassword(ctx context.Context, email, otp, newPassword string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil || user.OTP != otp || user.OTPExpires.Before(time.Now()) {
		return errors.New("invalid or expired OTP")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	user.OTP = ""
	user.OTPExpires = time.Time{}
	return s.userRepo.Update(ctx, user)
}