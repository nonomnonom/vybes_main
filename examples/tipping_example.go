package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"vybes/internal/domain"
	"vybes/internal/repository"
	"vybes/internal/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Example usage of the tipping service
func main() {
	// Connect to MongoDB (replace with your connection string)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("vybes")

	// Initialize repositories
	tippingRepo := repository.NewTippingRepository(db)
	
	// For this example, we'll create mock repositories
	// In a real application, you'd use actual repositories
	mockUserRepo := &MockUserRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockContentRepo := &MockContentRepository{}

	// Initialize tipping service
	tippingService := service.NewTippingService(
		tippingRepo,
		mockUserRepo,
		mockCommentRepo,
		mockContentRepo,
	)

	ctx := context.Background()

	// Example 1: Get or create allowance for a user
	fmt.Println("=== Example 1: Get/Create Allowance ===")
	userID := primitive.NewObjectID()
	allowance, err := tippingService.GetOrCreateAllowance(ctx, userID)
	if err != nil {
		log.Printf("Error getting allowance: %v", err)
	} else {
		fmt.Printf("User %s allowance: %d/%d VYB remaining\n", 
			userID.Hex(), 
			allowance.WeeklyLimit-allowance.UsedAmount, 
			allowance.WeeklyLimit)
	}

	// Example 2: Parse tip amount from comment text
	fmt.Println("\n=== Example 2: Parse Tip Amount ===")
	commentTexts := []string{
		"Great content! $100$vyb",
		"$250$vyb amazing work!",
		"Just a regular comment",
		"$abc$vyb invalid amount",
	}

	for _, text := range commentTexts {
		amount, err := tippingService.ParseTipAmount(text)
		if err != nil {
			fmt.Printf("Text: '%s' -> Error: %v\n", text, err)
		} else {
			fmt.Printf("Text: '%s' -> Amount: %d VYB\n", text, amount)
		}
	}

	// Example 3: Send a tip (with mock users)
	fmt.Println("\n=== Example 3: Send Tip ===")
	fromUserID := primitive.NewObjectID()
	toUserID := primitive.NewObjectID()

	// Setup mock users
	mockUserRepo.users = map[primitive.ObjectID]*domain.User{
		fromUserID: {ID: fromUserID, Name: "Sender"},
		toUserID:   {ID: toUserID, Name: "Receiver"},
	}

	// Create allowance for sender
	senderAllowance := &domain.TippingAllowance{
		ID:          primitive.NewObjectID(),
		UserID:      fromUserID,
		WeeklyLimit: 10000,
		UsedAmount:  0,
		WeekStart:   time.Now(),
		LastReset:   time.Now(),
	}
	tippingRepo.CreateAllowance(ctx, senderAllowance)

	// Send tip
	tip, err := tippingService.SendTip(ctx, fromUserID, toUserID, 500, "Great content!")
	if err != nil {
		log.Printf("Error sending tip: %v", err)
	} else {
		fmt.Printf("Tip sent successfully: %d VYB from %s to %s\n", 
			tip.Amount, 
			fromUserID.Hex(), 
			toUserID.Hex())
	}

	// Example 4: Process comment tip
	fmt.Println("\n=== Example 4: Process Comment Tip ===")
	commentID := primitive.NewObjectID()
	contentID := primitive.NewObjectID()
	commenterID := fromUserID
	creatorID := toUserID

	// Setup mock comment and content
	mockCommentRepo.comments = map[primitive.ObjectID]*domain.Comment{
		commentID: {
			ID:     commentID,
			UserID: commenterID,
			PostID: contentID,
			Text:   "Amazing content! $200$vyb",
		},
	}

	mockContentRepo.contents = map[primitive.ObjectID]*domain.Content{
		contentID: {
			ID:     contentID,
			UserID: creatorID,
		},
	}

	// Process comment tip
	err = tippingService.ProcessCommentTip(ctx, commentID)
	if err != nil {
		log.Printf("Error processing comment tip: %v", err)
	} else {
		fmt.Printf("Comment tip processed successfully: 200 VYB from commenter to creator\n")
	}

	// Example 5: Get tip statistics
	fmt.Println("\n=== Example 5: Get Tip Stats ===")
	stats, err := tippingService.GetTipStats(ctx, fromUserID)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
	} else {
		fmt.Printf("User %s stats:\n", fromUserID.Hex())
		fmt.Printf("  Total sent: %d VYB\n", stats.TotalSent)
		fmt.Printf("  Total received: %d VYB\n", stats.TotalReceived)
		fmt.Printf("  Weekly sent: %d VYB\n", stats.WeeklySent)
		fmt.Printf("  Weekly received: %d VYB\n", stats.WeeklyReceived)
	}

	fmt.Println("\n=== Tipping Service Example Complete ===")
}

// Mock repositories for example purposes
type MockUserRepository struct {
	users map[primitive.ObjectID]*domain.User
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

type MockCommentRepository struct {
	comments map[primitive.ObjectID]*domain.Comment
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error) {
	if comment, exists := m.comments[id]; exists {
		return comment, nil
	}
	return nil, fmt.Errorf("comment not found")
}

type MockContentRepository struct {
	contents map[primitive.ObjectID]*domain.Content
}

func (m *MockContentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Content, error) {
	if content, exists := m.contents[id]; exists {
		return content, nil
	}
	return nil, fmt.Errorf("content not found")
}