package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user data operations.
// Users are the core entities in the application, representing
// individuals who can create content, follow others, and interact
// with the platform.
type UserRepository interface {
	// CreateUser creates a new user account in the database
	CreateUser(ctx context.Context, user *domain.User) error
	// GetUserByID retrieves a user by their unique identifier
	GetUserByID(ctx context.Context, userID primitive.ObjectID) (*domain.User, error)
	// GetUserByEmail retrieves a user by their email address
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	// GetUserByUsername retrieves a user by their username
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	// GetUserByWalletAddress retrieves a user by their wallet address
	GetUserByWalletAddress(ctx context.Context, walletAddress string) (*domain.User, error)
	// UpdateUser updates an existing user's information
	UpdateUser(ctx context.Context, user *domain.User) error
	// DeleteUser removes a user account from the database
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
	// SearchUsers finds users based on search criteria (name, username)
	SearchUsers(ctx context.Context, query string, page, limit int) ([]domain.User, error)
	// GetUsersByIDs retrieves multiple users by their IDs
	GetUsersByIDs(ctx context.Context, userIDs []primitive.ObjectID) ([]domain.User, error)
}

// mongoUserRepository implements UserRepository using MongoDB as the backend
type mongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new user repository instance with MongoDB backend.
// The repository handles all user-related database operations including
// CRUD operations, authentication, and user search functionality.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - UserRepository: A configured user repository ready for use
func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		collection: db.Collection("users"),
	}
}
