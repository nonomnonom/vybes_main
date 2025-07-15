package main

import (
	"encoding/json"
	"fmt"
	"vybes/internal/domain"
	"vybes/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Example of the new LoginResponse structure
	userID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	user := &domain.User{
		ID:            userID,
		VID:           12345,
		Name:          "John Doe",
		Email:         "john@example.com",
		Username:      "johndoe",
		PFPURL:        "https://example.com/pfp.jpg",
		BannerURL:     "https://example.com/banner.jpg",
		Bio:           "Hello, I'm John!",
		WalletAddress: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
		TotalLikeCount: 100,
		PostCount:     50,
	}

	loginResponse := &service.LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
		UserData:     user,
	}

	// Convert to JSON to show the structure
	jsonData, err := json.MarshalIndent(loginResponse, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Println("New Login Response Structure:")
	fmt.Println(string(jsonData))
}