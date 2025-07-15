package main

import (
	"encoding/json"
	"fmt"
	"vybes/internal/service"
)

func main() {
	loginResponse := &service.LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
		UserData:     &service.UserMinimal{VID: 12345, Username: "johndoe"},
	}

	jsonData, err := json.MarshalIndent(loginResponse, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Println("New Login Response Structure:")
	fmt.Println(string(jsonData))
}
