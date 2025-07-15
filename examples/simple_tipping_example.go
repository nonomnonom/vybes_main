package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Simplified domain models for demonstration
type TippingAllowance struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	WeeklyLimit int64              `bson:"weeklyLimit" json:"weeklyLimit"`
	UsedAmount  int64              `bson:"usedAmount" json:"usedAmount"`
	WeekStart   time.Time          `bson:"weekStart" json:"weekStart"`
	LastReset   time.Time          `bson:"lastReset" json:"lastReset"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Tip struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	FromUserID  primitive.ObjectID  `bson:"fromUserId" json:"fromUserId"`
	ToUserID    primitive.ObjectID  `bson:"toUserId" json:"toUserId"`
	Amount      int64               `bson:"amount" json:"amount"`
	ContentID   *primitive.ObjectID `bson:"contentId,omitempty" json:"contentId,omitempty"`
	CommentID   *primitive.ObjectID `bson:"commentId,omitempty" json:"commentId,omitempty"`
	Message     string              `bson:"message,omitempty" json:"message,omitempty"`
	Status      string              `bson:"status" json:"status"`
	CreatedAt   time.Time           `bson:"createdAt" json:"createdAt"`
	CompletedAt *time.Time          `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
}

type TipStats struct {
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	TotalReceived  int64              `bson:"totalReceived" json:"totalReceived"`
	TotalSent      int64              `bson:"totalSent" json:"totalSent"`
	WeeklyReceived int64              `bson:"weeklyReceived" json:"weeklyReceived"`
	WeeklySent     int64              `bson:"weeklySent" json:"weeklySent"`
	LastUpdated    time.Time          `bson:"lastUpdated" json:"lastUpdated"`
}

// Simplified tipping service for demonstration
type SimpleTippingService struct {
	allowances map[primitive.ObjectID]*TippingAllowance
	tips       map[primitive.ObjectID]*Tip
	stats      map[primitive.ObjectID]*TipStats
}

func NewSimpleTippingService() *SimpleTippingService {
	return &SimpleTippingService{
		allowances: make(map[primitive.ObjectID]*TippingAllowance),
		tips:       make(map[primitive.ObjectID]*Tip),
		stats:      make(map[primitive.ObjectID]*TipStats),
	}
}

const (
	WeeklyAllowanceLimit = 10000
	TipPattern          = `\$(\d+)\$vyb`
)

// ParseTipAmount extracts tip amount from text like "$100$vyb"
func (s *SimpleTippingService) ParseTipAmount(text string) (int64, error) {
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

// GetOrCreateAllowance gets or creates a weekly allowance for a user
func (s *SimpleTippingService) GetOrCreateAllowance(userID primitive.ObjectID) (*TippingAllowance, error) {
	allowance, exists := s.allowances[userID]
	if !exists {
		now := time.Now()
		weekStart := s.getWeekStart(now)
		
		allowance = &TippingAllowance{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			WeeklyLimit: WeeklyAllowanceLimit,
			UsedAmount:  0,
			WeekStart:   weekStart,
			LastReset:   now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		
		s.allowances[userID] = allowance
	} else {
		// Check if we need to reset for new week
		if s.shouldResetAllowance(allowance) {
			allowance.UsedAmount = 0
			allowance.WeekStart = s.getWeekStart(time.Now())
			allowance.LastReset = time.Now()
			allowance.UpdatedAt = time.Now()
		}
	}
	
	return allowance, nil
}

// SendTip sends a tip from one user to another
func (s *SimpleTippingService) SendTip(fromUserID, toUserID primitive.ObjectID, amount int64, message string) (*Tip, error) {
	// Validate amount
	if amount <= 0 {
		return nil, fmt.Errorf("tip amount must be positive")
	}
	
	// Check allowance
	allowance, err := s.GetOrCreateAllowance(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}
	
	if allowance.UsedAmount+amount > allowance.WeeklyLimit {
		return nil, fmt.Errorf("insufficient weekly allowance. Used: %d, Limit: %d, Requested: %d", 
			allowance.UsedAmount, allowance.WeeklyLimit, amount)
	}
	
	// Create tip transaction
	tip := &Tip{
		ID:         primitive.NewObjectID(),
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Message:    message,
		Status:     "COMPLETED",
		CreatedAt:  time.Now(),
	}
	
	now := time.Now()
	tip.CompletedAt = &now
	
	// Store tip
	s.tips[tip.ID] = tip
	
	// Update allowance
	allowance.UsedAmount += amount
	allowance.UpdatedAt = time.Now()
	
	// Update stats
	s.updateTipStats(fromUserID, 0, amount)
	s.updateTipStats(toUserID, amount, 0)
	
	return tip, nil
}

// ProcessCommentTip processes a tip from a comment
func (s *SimpleTippingService) ProcessCommentTip(commentID, commenterID, contentID, creatorID primitive.ObjectID, commentText string) error {
	// Parse tip amount from comment text
	amount, err := s.ParseTipAmount(commentText)
	if err != nil {
		return fmt.Errorf("failed to parse tip amount: %w", err)
	}
	
	// Send tip from commenter to content creator
	tip, err := s.SendTip(commenterID, creatorID, amount, "Tip via comment on content")
	if err != nil {
		return fmt.Errorf("failed to send tip: %w", err)
	}
	
	// Update tip with content and comment references
	tip.ContentID = &contentID
	tip.CommentID = &commentID
	
	return nil
}

// GetTipStats gets tipping statistics for a user
func (s *SimpleTippingService) GetTipStats(userID primitive.ObjectID) *TipStats {
	stats, exists := s.stats[userID]
	if !exists {
		stats = &TipStats{
			UserID:      userID,
			LastUpdated: time.Now(),
		}
		s.stats[userID] = stats
	}
	return stats
}

// Helper functions
func (s *SimpleTippingService) getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	} else {
		weekday--
	}
	
	// Go back to Monday
	return t.AddDate(0, 0, -int(weekday)).Truncate(24 * time.Hour)
}

func (s *SimpleTippingService) shouldResetAllowance(allowance *TippingAllowance) bool {
	now := time.Now()
	currentWeekStart := s.getWeekStart(now)
	return allowance.WeekStart.Before(currentWeekStart)
}

func (s *SimpleTippingService) updateTipStats(userID primitive.ObjectID, received, sent int64) {
	stats, exists := s.stats[userID]
	if !exists {
		stats = &TipStats{
			UserID:      userID,
			LastUpdated: time.Now(),
		}
		s.stats[userID] = stats
	}
	
	stats.TotalReceived += received
	stats.TotalSent += sent
	stats.WeeklyReceived += received
	stats.WeeklySent += sent
	stats.LastUpdated = time.Now()
}

// Example usage
func main() {
	fmt.Println("=== Vybes Tipping Service Demo ===\n")

	// Initialize service
	tippingService := NewSimpleTippingService()

	// Example 1: Get or create allowance for a user
	fmt.Println("1. Get/Create Allowance")
	userID := primitive.NewObjectID()
	allowance, err := tippingService.GetOrCreateAllowance(userID)
	if err != nil {
		log.Printf("Error getting allowance: %v", err)
	} else {
		fmt.Printf("   User %s allowance: %d/%d VYB remaining\n", 
			userID.Hex()[:8], 
			allowance.WeeklyLimit-allowance.UsedAmount, 
			allowance.WeeklyLimit)
	}

	// Example 2: Parse tip amount from comment text
	fmt.Println("\n2. Parse Tip Amount from Comments")
	commentTexts := []string{
		"Great content! $100$vyb",
		"$250$vyb amazing work!",
		"Just a regular comment",
		"$abc$vyb invalid amount",
		"$0$vyb zero amount",
	}

	for _, text := range commentTexts {
		amount, err := tippingService.ParseTipAmount(text)
		if err != nil {
			fmt.Printf("   Text: '%s' -> Error: %v\n", text, err)
		} else {
			fmt.Printf("   Text: '%s' -> Amount: %d VYB\n", text, amount)
		}
	}

	// Example 3: Send a direct tip
	fmt.Println("\n3. Send Direct Tip")
	fromUserID := primitive.NewObjectID()
	toUserID := primitive.NewObjectID()

	tip, err := tippingService.SendTip(fromUserID, toUserID, 500, "Great content!")
	if err != nil {
		log.Printf("Error sending tip: %v", err)
	} else {
		fmt.Printf("   Tip sent successfully: %d VYB from %s to %s\n", 
			tip.Amount, 
			fromUserID.Hex()[:8], 
			toUserID.Hex()[:8])
	}

	// Example 4: Process comment tip
	fmt.Println("\n4. Process Comment Tip")
	commentID := primitive.NewObjectID()
	contentID := primitive.NewObjectID()
	commenterID := fromUserID
	creatorID := toUserID
	commentText := "Amazing content! $200$vyb"

	err = tippingService.ProcessCommentTip(commentID, commenterID, contentID, creatorID, commentText)
	if err != nil {
		log.Printf("Error processing comment tip: %v", err)
	} else {
		fmt.Printf("   Comment tip processed successfully: 200 VYB from commenter to creator\n")
	}

	// Example 5: Check allowance after tips
	fmt.Println("\n5. Check Allowance After Tips")
	updatedAllowance, _ := tippingService.GetOrCreateAllowance(fromUserID)
	fmt.Printf("   User %s allowance: %d/%d VYB remaining\n", 
		fromUserID.Hex()[:8], 
		updatedAllowance.WeeklyLimit-updatedAllowance.UsedAmount, 
		updatedAllowance.WeeklyLimit)

	// Example 6: Get tip statistics
	fmt.Println("\n6. Tip Statistics")
	stats := tippingService.GetTipStats(fromUserID)
	fmt.Printf("   User %s stats:\n", fromUserID.Hex()[:8])
	fmt.Printf("     Total sent: %d VYB\n", stats.TotalSent)
	fmt.Printf("     Total received: %d VYB\n", stats.TotalReceived)
	fmt.Printf("     Weekly sent: %d VYB\n", stats.WeeklySent)
	fmt.Printf("     Weekly received: %d VYB\n", stats.WeeklyReceived)

	// Example 7: Try to exceed allowance
	fmt.Println("\n7. Try to Exceed Allowance")
	largeTip, err := tippingService.SendTip(fromUserID, toUserID, 10000, "Large tip")
	if err != nil {
		fmt.Printf("   Error (expected): %v\n", err)
	} else {
		fmt.Printf("   Large tip sent: %d VYB\n", largeTip.Amount)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Weekly allowance management (10,000 VYB)")
	fmt.Println("✓ Tip pattern parsing ($amount$vyb)")
	fmt.Println("✓ Direct tipping between users")
	fmt.Println("✓ Comment-based tipping")
	fmt.Println("✓ Allowance enforcement")
	fmt.Println("✓ Statistics tracking")
	fmt.Println("✓ Automatic weekly reset logic")
}