package test

import (
	"context"
	"testing"
	"time"

	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Comment), args.Error(1)
}

type MockContentRepository struct {
	mock.Mock
}

func (m *MockContentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Content, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Content), args.Error(1)
}

type MockTippingRepository struct {
	mock.Mock
}

func (m *MockTippingRepository) CreateAllowance(ctx context.Context, allowance *domain.TippingAllowance) error {
	args := m.Called(ctx, allowance)
	return args.Error(0)
}

func (m *MockTippingRepository) GetAllowanceByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.TippingAllowance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TippingAllowance), args.Error(1)
}

func (m *MockTippingRepository) UpdateAllowance(ctx context.Context, allowance *domain.TippingAllowance) error {
	args := m.Called(ctx, allowance)
	return args.Error(0)
}

func (m *MockTippingRepository) CreateTip(ctx context.Context, tip *domain.Tip) error {
	args := m.Called(ctx, tip)
	return args.Error(0)
}

func (m *MockTippingRepository) UpdateTipStatus(ctx context.Context, tipID primitive.ObjectID, status domain.TipStatus) error {
	args := m.Called(ctx, tipID, status)
	return args.Error(0)
}

func (m *MockTippingRepository) UpdateTipStats(ctx context.Context, userID primitive.ObjectID, received, sent int64) error {
	args := m.Called(ctx, userID, received, sent)
	return args.Error(0)
}

func (m *MockTippingRepository) GetTipStats(ctx context.Context, userID primitive.ObjectID) (*domain.TipStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TipStats), args.Error(1)
}

func TestTippingService_ParseTipAmount(t *testing.T) {
	// Create mock repositories
	mockTippingRepo := &MockTippingRepository{}
	mockUserRepo := &MockUserRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockContentRepo := &MockContentRepository{}

	// Create service
	service := service.NewTippingService(mockTippingRepo, mockUserRepo, mockCommentRepo, mockContentRepo)

	tests := []struct {
		name    string
		text    string
		want    int64
		wantErr bool
	}{
		{
			name:    "valid tip amount",
			text:    "$100$vyb",
			want:    100,
			wantErr: false,
		},
		{
			name:    "valid tip amount with text",
			text:    "Great content! $250$vyb",
			want:    250,
			wantErr: false,
		},
		{
			name:    "multiple tips - should get first",
			text:    "$100$vyb and $200$vyb",
			want:    100,
			wantErr: false,
		},
		{
			name:    "no tip pattern",
			text:    "Just a regular comment",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid amount",
			text:    "$abc$vyb",
			want:    0,
			wantErr: true,
		},
		{
			name:    "zero amount",
			text:    "$0$vyb",
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative amount",
			text:    "$-100$vyb",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ParseTipAmount(tt.text)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTippingService_GetOrCreateAllowance(t *testing.T) {
	// Create mock repositories
	mockTippingRepo := &MockTippingRepository{}
	mockUserRepo := &MockUserRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockContentRepo := &MockContentRepository{}

	// Create service
	service := service.NewTippingService(mockTippingRepo, mockUserRepo, mockCommentRepo, mockContentRepo)

	ctx := context.Background()
	userID := primitive.NewObjectID()

	t.Run("create new allowance", func(t *testing.T) {
		// Mock that allowance doesn't exist
		mockTippingRepo.On("GetAllowanceByUserID", ctx, userID).Return(nil, assert.AnError)

		// Mock allowance creation
		mockTippingRepo.On("CreateAllowance", ctx, mock.AnythingOfType("*domain.TippingAllowance")).Return(nil)

		allowance, err := service.GetOrCreateAllowance(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, allowance)
		assert.Equal(t, userID, allowance.UserID)
		assert.Equal(t, int64(10000), allowance.WeeklyLimit)
		assert.Equal(t, int64(0), allowance.UsedAmount)

		mockTippingRepo.AssertExpectations(t)
	})

	t.Run("get existing allowance", func(t *testing.T) {
		existingAllowance := &domain.TippingAllowance{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			WeeklyLimit: 10000,
			UsedAmount:  2500,
			WeekStart:   time.Now().AddDate(0, 0, -1), // Yesterday
			LastReset:   time.Now().AddDate(0, 0, -1),
		}

		// Mock that allowance exists
		mockTippingRepo.On("GetAllowanceByUserID", ctx, userID).Return(existingAllowance, nil)

		allowance, err := service.GetOrCreateAllowance(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, allowance)
		assert.Equal(t, userID, allowance.UserID)
		assert.Equal(t, int64(10000), allowance.WeeklyLimit)
		assert.Equal(t, int64(2500), allowance.UsedAmount)

		mockTippingRepo.AssertExpectations(t)
	})
}

func TestTippingService_SendTip(t *testing.T) {
	// Create mock repositories
	mockTippingRepo := &MockTippingRepository{}
	mockUserRepo := &MockUserRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockContentRepo := &MockContentRepository{}

	// Create service
	service := service.NewTippingService(mockTippingRepo, mockUserRepo, mockCommentRepo, mockContentRepo)

	ctx := context.Background()
	fromUserID := primitive.NewObjectID()
	toUserID := primitive.NewObjectID()

	fromUser := &domain.User{ID: fromUserID, Name: "Sender"}
	toUser := &domain.User{ID: toUserID, Name: "Receiver"}

	allowance := &domain.TippingAllowance{
		ID:          primitive.NewObjectID(),
		UserID:      fromUserID,
		WeeklyLimit: 10000,
		UsedAmount:  0,
		WeekStart:   time.Now(),
		LastReset:   time.Now(),
	}

	t.Run("successful tip", func(t *testing.T) {
		// Mock user lookups
		mockUserRepo.On("GetByID", ctx, fromUserID).Return(fromUser, nil)
		mockUserRepo.On("GetByID", ctx, toUserID).Return(toUser, nil)

		// Mock allowance
		mockTippingRepo.On("GetAllowanceByUserID", ctx, fromUserID).Return(allowance, nil)

		// Mock tip creation
		mockTippingRepo.On("CreateTip", ctx, mock.AnythingOfType("*domain.Tip")).Return(nil)

		// Mock allowance update
		mockTippingRepo.On("UpdateAllowance", ctx, mock.AnythingOfType("*domain.TippingAllowance")).Return(nil)

		// Mock stats updates
		mockTippingRepo.On("UpdateTipStats", ctx, fromUserID, int64(0), int64(100)).Return(nil)
		mockTippingRepo.On("UpdateTipStats", ctx, toUserID, int64(100), int64(0)).Return(nil)

		// Mock tip status update
		mockTippingRepo.On("UpdateTipStatus", ctx, mock.AnythingOfType("primitive.ObjectID"), domain.TipStatusCompleted).Return(nil)

		tip, err := service.SendTip(ctx, fromUserID, toUserID, 100, "Great content!")

		assert.NoError(t, err)
		assert.NotNil(t, tip)
		assert.Equal(t, fromUserID, tip.FromUserID)
		assert.Equal(t, toUserID, tip.ToUserID)
		assert.Equal(t, int64(100), tip.Amount)
		assert.Equal(t, "Great content!", tip.Message)
		assert.Equal(t, domain.TipStatusCompleted, tip.Status)

		mockTippingRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("insufficient allowance", func(t *testing.T) {
		// Create allowance with insufficient balance
		insufficientAllowance := &domain.TippingAllowance{
			ID:          primitive.NewObjectID(),
			UserID:      fromUserID,
			WeeklyLimit: 10000,
			UsedAmount:  9500, // Only 500 remaining
			WeekStart:   time.Now(),
			LastReset:   time.Now(),
		}

		// Mock user lookups
		mockUserRepo.On("GetByID", ctx, fromUserID).Return(fromUser, nil)
		mockUserRepo.On("GetByID", ctx, toUserID).Return(toUser, nil)

		// Mock allowance
		mockTippingRepo.On("GetAllowanceByUserID", ctx, fromUserID).Return(insufficientAllowance, nil)

		tip, err := service.SendTip(ctx, fromUserID, toUserID, 1000, "Great content!")

		assert.Error(t, err)
		assert.Nil(t, tip)
		assert.Contains(t, err.Error(), "insufficient weekly allowance")

		mockTippingRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid amount", func(t *testing.T) {
		tip, err := service.SendTip(ctx, fromUserID, toUserID, 0, "Great content!")

		assert.Error(t, err)
		assert.Nil(t, tip)
		assert.Contains(t, err.Error(), "tip amount must be positive")
	})
}

func TestTippingService_ProcessCommentTip(t *testing.T) {
	// Create mock repositories
	mockTippingRepo := &MockTippingRepository{}
	mockUserRepo := &MockUserRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockContentRepo := &MockContentRepository{}

	// Create service
	service := service.NewTippingService(mockTippingRepo, mockUserRepo, mockCommentRepo, mockContentRepo)

	ctx := context.Background()
	commentID := primitive.NewObjectID()
	commenterID := primitive.NewObjectID()
	contentID := primitive.NewObjectID()
	creatorID := primitive.NewObjectID()

	comment := &domain.Comment{
		ID:     commentID,
		UserID: commenterID,
		PostID: contentID,
		Text:   "Great content! $100$vyb",
	}

	content := &domain.Content{
		ID:     contentID,
		UserID: creatorID,
	}

	commenter := &domain.User{ID: commenterID, Name: "Commenter"}
	creator := &domain.User{ID: creatorID, Name: "Creator"}

	allowance := &domain.TippingAllowance{
		ID:          primitive.NewObjectID(),
		UserID:      commenterID,
		WeeklyLimit: 10000,
		UsedAmount:  0,
		WeekStart:   time.Now(),
		LastReset:   time.Now(),
	}

	t.Run("successful comment tip", func(t *testing.T) {
		// Mock comment lookup
		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		// Mock content lookup
		mockContentRepo.On("GetByID", ctx, contentID).Return(content, nil)

		// Mock user lookups
		mockUserRepo.On("GetByID", ctx, commenterID).Return(commenter, nil)
		mockUserRepo.On("GetByID", ctx, creatorID).Return(creator, nil)

		// Mock allowance
		mockTippingRepo.On("GetAllowanceByUserID", ctx, commenterID).Return(allowance, nil)

		// Mock tip creation
		mockTippingRepo.On("CreateTip", ctx, mock.AnythingOfType("*domain.Tip")).Return(nil)

		// Mock allowance update
		mockTippingRepo.On("UpdateAllowance", ctx, mock.AnythingOfType("*domain.TippingAllowance")).Return(nil)

		// Mock stats updates
		mockTippingRepo.On("UpdateTipStats", ctx, commenterID, int64(0), int64(100)).Return(nil)
		mockTippingRepo.On("UpdateTipStats", ctx, creatorID, int64(100), int64(0)).Return(nil)

		// Mock tip status updates
		mockTippingRepo.On("UpdateTipStatus", ctx, mock.AnythingOfType("primitive.ObjectID"), domain.TipStatusCompleted).Return(nil).Times(2)

		err := service.ProcessCommentTip(ctx, commentID)

		assert.NoError(t, err)

		mockTippingRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
		mockContentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		// Mock comment not found
		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, assert.AnError)

		err := service.ProcessCommentTip(ctx, commentID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "comment not found")

		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("invalid tip pattern", func(t *testing.T) {
		invalidComment := &domain.Comment{
			ID:     commentID,
			UserID: commenterID,
			PostID: contentID,
			Text:   "Just a regular comment without tip",
		}

		// Mock comment lookup
		mockCommentRepo.On("GetByID", ctx, commentID).Return(invalidComment, nil)

		err := service.ProcessCommentTip(ctx, commentID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid tip pattern found")

		mockCommentRepo.AssertExpectations(t)
	})
}