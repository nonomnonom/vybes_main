package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/pkg/evm"
	"vybes/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WalletSecurityService defines the interface for wallet security operations.
type WalletSecurityService interface {
	CreateWalletSession(ctx context.Context, userID string, password string, ipAddress, userAgent string) (*domain.WalletSession, error)
	ValidateWalletSession(ctx context.Context, sessionToken string) (*domain.WalletSession, error)
	RevokeWalletSession(ctx context.Context, sessionToken string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	GetNextNonce(ctx context.Context, userID string) (uint64, error)
	ValidateAndUpdateNonce(ctx context.Context, userID string, txNonce uint64) error
	LogWalletAction(ctx context.Context, userID string, action string, txHash string, nonce uint64, amount, toAddress, ipAddress, userAgent, status, errorMsg string) error
	GetWalletAuditLogs(ctx context.Context, userID string, limit int64) ([]*domain.WalletAuditLog, error)
	PersonalSignWithSession(ctx context.Context, sessionToken, message string) (string, error)
	SignTransactionWithSession(ctx context.Context, sessionToken string, tx *types.Transaction) (*types.Transaction, error)
	SendTransactionWithSession(ctx context.Context, sessionToken string, tx *types.Transaction) (common.Hash, error)
	SignTypedDataV4WithSession(ctx context.Context, sessionToken string, typedData apitypes.TypedData) (string, error)
	Secp256k1SignWithSession(ctx context.Context, sessionToken, hash string) (string, error)
}

type walletSecurityService struct {
	userRepo            repository.UserRepository
	walletSessionRepo   repository.WalletSessionRepository
	walletAuditRepo     repository.WalletAuditRepository
	walletService       WalletService
	jwtSecret           string
	walletEncryptionKey string
}

// NewWalletSecurityService creates a new wallet security service.
func NewWalletSecurityService(
	userRepo repository.UserRepository,
	walletSessionRepo repository.WalletSessionRepository,
	walletAuditRepo repository.WalletAuditRepository,
	walletService WalletService,
	jwtSecret, walletEncryptionKey string,
) WalletSecurityService {
	return &walletSecurityService{
		userRepo:            userRepo,
		walletSessionRepo:   walletSessionRepo,
		walletAuditRepo:     walletAuditRepo,
		walletService:       walletService,
		jwtSecret:           jwtSecret,
		walletEncryptionKey: walletEncryptionKey,
	}
}

// CreateWalletSession creates a temporary wallet access session.
func (s *walletSecurityService) CreateWalletSession(ctx context.Context, userID string, password string, ipAddress, userAgent string) (*domain.WalletSession, error) {
	// Parse user ID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get user and validate password
	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if account is locked
	if !user.AccountLockedUntil.IsZero() && user.AccountLockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("account is locked until %s", user.AccountLockedUntil.Format(time.RFC3339))
	}

	// Validate password
	if !utils.CheckPasswordHash(password, user.Password) {
		// Increment failed attempts
		s.userRepo.IncrementFailedLoginAttempts(ctx, objID)
		
		// Lock account if too many failed attempts
		if user.FailedLoginAttempts >= 4 {
			lockUntil := time.Now().Add(15 * time.Minute)
			s.userRepo.LockAccount(ctx, objID, lockUntil)
			return nil, fmt.Errorf("account locked due to too many failed attempts. Try again after %s", lockUntil.Format(time.RFC3339))
		}
		
		return nil, errors.New("invalid password")
	}

	// Reset failed login attempts on successful login
	s.userRepo.ResetFailedLoginAttempts(ctx, objID)

	// Generate session token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return nil, err
	}

	// Create session
	session := &domain.WalletSession{
		ID:           primitive.NewObjectID(),
		UserID:       objID,
		SessionToken: sessionToken,
		ExpiresAt:    time.Now().Add(30 * time.Minute), // 30 minutes session
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	if err := s.walletSessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Update last wallet access
	s.userRepo.UpdateLastWalletAccess(ctx, objID)

	// Log the session creation
	s.LogWalletAction(ctx, userID, "session_created", "", 0, "", "", ipAddress, userAgent, "success", "")

	return session, nil
}

// ValidateWalletSession validates a wallet session token.
func (s *walletSecurityService) ValidateWalletSession(ctx context.Context, sessionToken string) (*domain.WalletSession, error) {
	session, err := s.walletSessionRepo.FindByToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("invalid session token")
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		// Delete expired session
		s.walletSessionRepo.Delete(ctx, session.ID)
		return nil, errors.New("session expired")
	}

	// Update last used timestamp
	s.walletSessionRepo.UpdateLastUsed(ctx, session.ID)

	return session, nil
}

// RevokeWalletSession revokes a specific wallet session.
func (s *walletSecurityService) RevokeWalletSession(ctx context.Context, sessionToken string) error {
	session, err := s.walletSessionRepo.FindByToken(ctx, sessionToken)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	return s.walletSessionRepo.Delete(ctx, session.ID)
}

// RevokeAllUserSessions revokes all sessions for a user.
func (s *walletSecurityService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return s.walletSessionRepo.DeleteByUserID(ctx, objID)
}

// GetNextNonce gets the next nonce for a user's wallet.
func (s *walletSecurityService) GetNextNonce(ctx context.Context, userID string) (uint64, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, errors.New("user not found")
	}

	return user.WalletNonce, nil
}

// ValidateAndUpdateNonce validates and updates the nonce for a transaction.
func (s *walletSecurityService) ValidateAndUpdateNonce(ctx context.Context, userID string, txNonce uint64) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	user, err := s.userRepo.FindByID(ctx, objID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Validate nonce
	if txNonce != user.WalletNonce {
		return fmt.Errorf("invalid nonce. Expected %d, got %d", user.WalletNonce, txNonce)
	}

	// Update nonce
	return s.userRepo.UpdateWalletNonce(ctx, objID, user.WalletNonce+1)
}

// LogWalletAction logs a wallet action for audit purposes.
func (s *walletSecurityService) LogWalletAction(ctx context.Context, userID string, action string, txHash string, nonce uint64, amount, toAddress, ipAddress, userAgent, status, errorMsg string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	audit := &domain.WalletAuditLog{
		ID:        primitive.NewObjectID(),
		UserID:    objID,
		Action:    action,
		TxHash:    txHash,
		Nonce:     nonce,
		Amount:    amount,
		ToAddress: toAddress,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Status:    status,
		ErrorMsg:  errorMsg,
		CreatedAt: time.Now(),
	}

	return s.walletAuditRepo.Create(ctx, audit)
}

// GetWalletAuditLogs gets audit logs for a user.
func (s *walletSecurityService) GetWalletAuditLogs(ctx context.Context, userID string, limit int64) ([]*domain.WalletAuditLog, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	return s.walletAuditRepo.FindByUserID(ctx, objID, limit)
}

// PersonalSignWithSession signs a message using a wallet session.
func (s *walletSecurityService) PersonalSignWithSession(ctx context.Context, sessionToken, message string) (string, error) {
	session, err := s.ValidateWalletSession(ctx, sessionToken)
	if err != nil {
		return "", err
	}

	// Get user's private key
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", err
	}

	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return "", errors.New("could not decrypt private key")
	}

	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}

	signature, err := signer.PersonalSign([]byte(message))
	if err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "personal_sign", "", 0, "", "", "", "", "error", err.Error())
		return "", err
	}

	s.LogWalletAction(ctx, session.UserID.Hex(), "personal_sign", "", 0, "", "", "", "", "success", "")
	return signature, nil
}

// SignTransactionWithSession signs a transaction using a wallet session.
func (s *walletSecurityService) SignTransactionWithSession(ctx context.Context, sessionToken string, tx *types.Transaction) (*types.Transaction, error) {
	session, err := s.ValidateWalletSession(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// Validate and update nonce
	if err := s.ValidateAndUpdateNonce(ctx, session.UserID.Hex(), tx.Nonce()); err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "sign_transaction", "", tx.Nonce(), "", "", "", "", "error", err.Error())
		return nil, err
	}

	// Get user's private key
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return nil, errors.New("could not decrypt private key")
	}

	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return nil, err
	}

	signedTx, err := signer.SignTransaction(tx)
	if err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "sign_transaction", "", tx.Nonce(), "", "", "", "", "error", err.Error())
		return nil, err
	}

	s.LogWalletAction(ctx, session.UserID.Hex(), "sign_transaction", "", tx.Nonce(), "", "", "", "", "success", "")
	return signedTx, nil
}

// SendTransactionWithSession sends a transaction using a wallet session.
func (s *walletSecurityService) SendTransactionWithSession(ctx context.Context, sessionToken string, tx *types.Transaction) (common.Hash, error) {
	session, err := s.ValidateWalletSession(ctx, sessionToken)
	if err != nil {
		return common.Hash{}, err
	}

	// Validate and update nonce
	if err := s.ValidateAndUpdateNonce(ctx, session.UserID.Hex(), tx.Nonce()); err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "send_transaction", "", tx.Nonce(), "", "", "", "", "error", err.Error())
		return common.Hash{}, err
	}

	// Get user's private key
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return common.Hash{}, err
	}

	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return common.Hash{}, errors.New("could not decrypt private key")
	}

	txHash, err := s.walletService.SendTransaction(ctx, tx, decryptedKey)
	if err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "send_transaction", "", tx.Nonce(), "", "", "", "", "error", err.Error())
		return common.Hash{}, err
	}

	s.LogWalletAction(ctx, session.UserID.Hex(), "send_transaction", txHash.Hex(), tx.Nonce(), "", "", "", "", "success", "")
	return txHash, nil
}

// SignTypedDataV4WithSession signs typed data using a wallet session.
func (s *walletSecurityService) SignTypedDataV4WithSession(ctx context.Context, sessionToken string, typedData apitypes.TypedData) (string, error) {
	session, err := s.ValidateWalletSession(ctx, sessionToken)
	if err != nil {
		return "", err
	}

	// Get user's private key
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", err
	}

	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return "", errors.New("could not decrypt private key")
	}

	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}

	signature, err := signer.SignTypedDataV4(typedData)
	if err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "sign_typed_data_v4", "", 0, "", "", "", "", "error", err.Error())
		return "", err
	}

	s.LogWalletAction(ctx, session.UserID.Hex(), "sign_typed_data_v4", "", 0, "", "", "", "", "success", "")
	return signature, nil
}

// Secp256k1SignWithSession signs a hash using a wallet session.
func (s *walletSecurityService) Secp256k1SignWithSession(ctx context.Context, sessionToken, hash string) (string, error) {
	session, err := s.ValidateWalletSession(ctx, sessionToken)
	if err != nil {
		return "", err
	}

	// Get user's private key
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", err
	}

	decryptedKey, err := utils.Decrypt(user.EncryptedPrivateKey, s.walletEncryptionKey)
	if err != nil {
		return "", errors.New("could not decrypt private key")
	}

	signer, err := evm.NewSignerFromHex(decryptedKey)
	if err != nil {
		return "", err
	}

	hashBytes, err := hexutil.Decode(hash)
	if err != nil {
		return "", errors.New("invalid hash format")
	}

	signature, err := signer.Secp256k1Sign(hashBytes)
	if err != nil {
		s.LogWalletAction(ctx, session.UserID.Hex(), "secp256k1_sign", "", 0, "", "", "", "", "error", err.Error())
		return "", err
	}

	s.LogWalletAction(ctx, session.UserID.Hex(), "secp256k1_sign", "", 0, "", "", "", "", "success", "")
	return signature, nil
}

// generateSessionToken generates a cryptographically secure session token.
func (s *walletSecurityService) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}