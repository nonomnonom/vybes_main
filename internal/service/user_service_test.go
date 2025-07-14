package service

import (
	"context"
	"testing"
	"time"
	"vybes/internal/domain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) { return nil, nil }
func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) { return nil, nil }
func (m *MockUserRepository) FindManyByIDs(ctx context.Context, ids []primitive.ObjectID) ([]domain.User, error) { return nil, nil }
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *MockUserRepository) IncrementTotalLikes(ctx context.Context, userID primitive.ObjectID, amount int) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}
func (m *MockUserRepository) IncrementPostCount(ctx context.Context, userID primitive.ObjectID, amount int) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}
func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}


// MockCounterRepository is a mock for CounterRepository
type MockCounterRepository struct {
	mock.Mock
}

func (m *MockCounterRepository) GetNextSequence(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(int64), args.Error(1)
}

// MockWalletService is a mock for WalletService
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}
func (m *MockWalletService) SendTransaction(ctx context.Context, tx *types.Transaction, privateKeyHex string) (common.Hash, error) {
	args := m.Called(ctx, tx, privateKeyHex)
	return args.Get(0).(common.Hash), args.Error(1)
}

// MockCache is a mock for cache.Client
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}
func (m *MockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}


func TestRegister(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockCounterRepo := new(MockCounterRepository)
	mockWalletService := new(MockWalletService)
	mockCache := new(MockCache)
	
	// We can use a nil for dependencies not used in this specific test
	userService := NewUserService(mockUserRepo, nil, mockCounterRepo, mockWalletService, nil, mockCache, "", "")

	ctx := context.Background()
	testEmail := "test@example.com"

	t.Run("success", func(t *testing.T) {
		// Setup expectations
		mockUserRepo.On("FindByEmail", ctx, testEmail).Return(nil, nil).Once()
		mockCounterRepo.On("GetNextSequence", ctx, "user_vid").Return(int64(1), nil).Once()
		mockWalletService.On("CreateWallet").Return("0xaddress", "encrypted_key", nil).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		// Execute the service method
		err := userService.Register(ctx, "Test User", testEmail, "password123")

		// Assertions
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockCounterRepo.AssertExpectations(t)
		mockWalletService.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		// Setup expectations
		existingUser := &domain.User{Email: testEmail}
		mockUserRepo.On("FindByEmail", ctx, testEmail).Return(existingUser, nil).Once()

		// Execute the service method
		err := userService.Register(ctx, "Test User", testEmail, "password123")

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, "user with this email already exists", err.Error())
		mockUserRepo.AssertExpectations(t)
	})
}