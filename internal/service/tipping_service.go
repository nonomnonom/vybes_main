package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"vybes/internal/domain"
	"vybes/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TippingService struct {
	tippingRepo *repository.TippingRepository
	userRepo    *repository.UserRepository
	commentRepo *repository.CommentRepository
	contentRepo *repository.ContentRepository
}

func NewTippingService(
	tippingRepo *repository.TippingRepository,
	userRepo *repository.UserRepository,
	commentRepo *repository.CommentRepository,
	contentRepo *repository.ContentRepository,
) *TippingService {
	return &TippingService{
		tippingRepo: tippingRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
		contentRepo: contentRepo,
	}
}

const (
	WeeklyAllowanceLimit = 10000 // 10k VYB per week
	TipPattern          = `\$(\d+)\$vyb` // Pattern to match $100$vyb
)

// GetOrCreateAllowance gets or creates a weekly allowance for a user
func (s *TippingService) GetOrCreateAllowance(ctx context.Context, userID primitive.ObjectID) (*domain.TippingAllowance, error) {
	allowance, err := s.tippingRepo.GetAllowanceByUserID(ctx, userID)
	if err != nil {
		// Create new allowance if not exists
		now := time.Now()
		weekStart := s.getWeekStart(now)
		
		allowance = &domain.TippingAllowance{
			UserID:      userID,
			WeeklyLimit: WeeklyAllowanceLimit,
			UsedAmount:  0,
			WeekStart:   weekStart,
			LastReset:   now,
		}
		
		err = s.tippingRepo.CreateAllowance(ctx, allowance)
		if err != nil {
			return nil, fmt.Errorf("failed to create allowance: %w", err)
		}
	} else {
		// Check if we need to reset for new week
		if s.shouldResetAllowance(allowance) {
			allowance.UsedAmount = 0
			allowance.WeekStart = s.getWeekStart(time.Now())
			allowance.LastReset = time.Now()
			
			err = s.tippingRepo.UpdateAllowance(ctx, allowance)
			if err != nil {
				return nil, fmt.Errorf("failed to reset allowance: %w", err)
			}
		}
	}
	
	return allowance, nil
}

// SendTip sends a tip from one user to another
func (s *TippingService) SendTip(ctx context.Context, fromUserID, toUserID primitive.ObjectID, amount int64, message string) (*domain.Tip, error) {
	// Validate amount
	if amount <= 0 {
		return nil, fmt.Errorf("tip amount must be positive")
	}
	
	// Check if users exist
	fromUser, err := s.userRepo.GetByID(ctx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("sender user not found: %w", err)
	}
	
	toUser, err := s.userRepo.GetByID(ctx, toUserID)
	if err != nil {
		return nil, fmt.Errorf("recipient user not found: %w", err)
	}
	
	// Check allowance
	allowance, err := s.GetOrCreateAllowance(ctx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}
	
	if allowance.UsedAmount+amount > allowance.WeeklyLimit {
		return nil, fmt.Errorf("insufficient weekly allowance. Used: %d, Limit: %d, Requested: %d", 
			allowance.UsedAmount, allowance.WeeklyLimit, amount)
	}
	
	// Create tip transaction
	tip := &domain.Tip{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Message:    message,
		Status:     domain.TipStatusPending,
	}
	
	err = s.tippingRepo.CreateTip(ctx, tip)
	if err != nil {
		return nil, fmt.Errorf("failed to create tip: %w", err)
	}
	
	// Update allowance
	allowance.UsedAmount += amount
	err = s.tippingRepo.UpdateAllowance(ctx, allowance)
	if err != nil {
		return nil, fmt.Errorf("failed to update allowance: %w", err)
	}
	
	// Update stats
	err = s.tippingRepo.UpdateTipStats(ctx, fromUserID, 0, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to update sender stats: %w", err)
	}
	
	err = s.tippingRepo.UpdateTipStats(ctx, toUserID, amount, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipient stats: %w", err)
	}
	
	// Mark tip as completed
	err = s.tippingRepo.UpdateTipStatus(ctx, tip.ID, domain.TipStatusCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to complete tip: %w", err)
	}
	
	tip.Status = domain.TipStatusCompleted
	now := time.Now()
	tip.CompletedAt = &now
	
	return tip, nil
}

// ProcessCommentTip processes a tip from a comment
func (s *TippingService) ProcessCommentTip(ctx context.Context, commentID primitive.ObjectID) error {
	// Get comment
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}
	
	// Parse tip amount from comment text
	amount, err := s.parseTipAmount(comment.Text)
	if err != nil {
		return fmt.Errorf("failed to parse tip amount: %w", err)
	}
	
	// Get content to find the content creator
	content, err := s.contentRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		return fmt.Errorf("content not found: %w", err)
	}
	
	// Send tip from commenter to content creator
	tip, err := s.SendTip(ctx, comment.UserID, content.UserID, amount, fmt.Sprintf("Tip via comment on content"))
	if err != nil {
		return fmt.Errorf("failed to send tip: %w", err)
	}
	
	// Update tip with content and comment references
	tip.ContentID = &content.ID
	tip.CommentID = &comment.ID
	
	err = s.tippingRepo.UpdateTipStatus(ctx, tip.ID, domain.TipStatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to update tip with content reference: %w", err)
	}
	
	return nil
}

// ParseTipAmount extracts tip amount from text like "$100$vyb"
func (s *TippingService) parseTipAmount(text string) (int64, error) {
	re := regexp.MustCompile(TipPattern)
	matches := re.FindStringSubmatch(text)
	
	if len(matches) < 2 {
		return 0, fmt.Errorf("no valid tip pattern found in text")
	}
	
	amount, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid tip amount: %w", err)
	}
	
	if amount <= 0 {
		return 0, fmt.Errorf("tip amount must be positive")
	}
	
	return amount, nil
}

// GetUserTips gets all tips for a user (sent and received)
func (s *TippingService) GetUserTips(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*domain.Tip, error) {
	return s.tippingRepo.GetTipsByUser(ctx, userID, limit)
}

// GetContentTips gets all tips for a specific content
func (s *TippingService) GetContentTips(ctx context.Context, contentID primitive.ObjectID) ([]*domain.Tip, error) {
	return s.tippingRepo.GetTipsByContent(ctx, contentID)
}

// GetTipStats gets tipping statistics for a user
func (s *TippingService) GetTipStats(ctx context.Context, userID primitive.ObjectID) (*domain.TipStats, error) {
	return s.tippingRepo.GetTipStats(ctx, userID)
}

// ResetWeeklyAllowances resets all weekly allowances (called by cron job)
func (s *TippingService) ResetWeeklyAllowances(ctx context.Context) error {
	return s.tippingRepo.ResetWeeklyAllowances(ctx)
}

// ResetWeeklyStats resets all weekly stats (called by cron job)
func (s *TippingService) ResetWeeklyStats(ctx context.Context) error {
	return s.tippingRepo.ResetWeeklyStats(ctx)
}

// Helper functions
func (s *TippingService) getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	} else {
		weekday--
	}
	
	// Go back to Monday
	return t.AddDate(0, 0, -int(weekday)).Truncate(24 * time.Hour)
}

func (s *TippingService) shouldResetAllowance(allowance *domain.TippingAllowance) bool {
	now := time.Now()
	currentWeekStart := s.getWeekStart(now)
	return allowance.WeekStart.Before(currentWeekStart)
}