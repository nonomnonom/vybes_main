package repository

import (
	"context"
	"time"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	IncrementFailedLoginAttempts(ctx context.Context, userID primitive.ObjectID) error
	ResetFailedLoginAttempts(ctx context.Context, userID primitive.ObjectID) error
	LockAccount(ctx context.Context, userID primitive.ObjectID, lockUntil time.Time) error
	UnlockAccount(ctx context.Context, userID primitive.ObjectID) error
	UpdateWalletNonce(ctx context.Context, userID primitive.ObjectID, nonce uint64) error
	UpdateLastWalletAccess(ctx context.Context, userID primitive.ObjectID) error
}

// WalletSessionRepository defines the interface for wallet session operations.
type WalletSessionRepository interface {
	Create(ctx context.Context, session *domain.WalletSession) error
	FindByToken(ctx context.Context, token string) (*domain.WalletSession, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.WalletSession, error)
	UpdateLastUsed(ctx context.Context, sessionID primitive.ObjectID) error
	Delete(ctx context.Context, sessionID primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

// WalletAuditRepository defines the interface for wallet audit operations.
type WalletAuditRepository interface {
	Create(ctx context.Context, audit *domain.WalletAuditLog) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*domain.WalletAuditLog, error)
	FindByTxHash(ctx context.Context, txHash string) (*domain.WalletAuditLog, error)
}

type userRepository struct {
	collection *mongo.Collection
}

type walletSessionRepository struct {
	collection *mongo.Collection
}

type walletAuditRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

// NewWalletSessionRepository creates a new wallet session repository.
func NewWalletSessionRepository(db *mongo.Database) WalletSessionRepository {
	return &walletSessionRepository{
		collection: db.Collection("wallet_sessions"),
	}
}

// NewWalletAuditRepository creates a new wallet audit repository.
func NewWalletAuditRepository(db *mongo.Database) WalletAuditRepository {
	return &walletAuditRepository{
		collection: db.Collection("wallet_audit_logs"),
	}
}

// User Repository Methods
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	return err
}

func (r *userRepository) IncrementFailedLoginAttempts(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"failedLoginAttempts": 1}},
	)
	return err
}

func (r *userRepository) ResetFailedLoginAttempts(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"failedLoginAttempts": 0}},
	)
	return err
}

func (r *userRepository) LockAccount(ctx context.Context, userID primitive.ObjectID, lockUntil time.Time) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"accountLockedUntil": lockUntil}},
	)
	return err
}

func (r *userRepository) UnlockAccount(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$unset": bson.M{"accountLockedUntil": ""}},
	)
	return err
}

func (r *userRepository) UpdateWalletNonce(ctx context.Context, userID primitive.ObjectID, nonce uint64) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"walletNonce": nonce}},
	)
	return err
}

func (r *userRepository) UpdateLastWalletAccess(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"lastWalletAccess": time.Now()}},
	)
	return err
}

// Wallet Session Repository Methods
func (r *walletSessionRepository) Create(ctx context.Context, session *domain.WalletSession) error {
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *walletSessionRepository) FindByToken(ctx context.Context, token string) (*domain.WalletSession, error) {
	var session domain.WalletSession
	err := r.collection.FindOne(ctx, bson.M{"sessionToken": token}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *walletSessionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.WalletSession, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*domain.WalletSession
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *walletSessionRepository) UpdateLastUsed(ctx context.Context, sessionID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": sessionID},
		bson.M{"$set": bson.M{"lastUsedAt": time.Now()}},
	)
	return err
}

func (r *walletSessionRepository) Delete(ctx context.Context, sessionID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": sessionID})
	return err
}

func (r *walletSessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"expiresAt": bson.M{"$lt": time.Now()}})
	return err
}

func (r *walletSessionRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"userId": userID})
	return err
}

// Wallet Audit Repository Methods
func (r *walletAuditRepository) Create(ctx context.Context, audit *domain.WalletAuditLog) error {
	_, err := r.collection.InsertOne(ctx, audit)
	return err
}

func (r *walletAuditRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*domain.WalletAuditLog, error) {
	opts := options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var audits []*domain.WalletAuditLog
	if err = cursor.All(ctx, &audits); err != nil {
		return nil, err
	}
	return audits, nil
}

func (r *walletAuditRepository) FindByTxHash(ctx context.Context, txHash string) (*domain.WalletAuditLog, error) {
	var audit domain.WalletAuditLog
	err := r.collection.FindOne(ctx, bson.M{"txHash": txHash}).Decode(&audit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &audit, nil
}