package service

import (
	"context"
	"log"
	"math/rand"
	"time"
)

// UserClient defines the interface for user operations
type UserClient interface {
	GetUserById(ctx context.Context, userId string) (*UserInfo, error)
}

// UserInfo represents user information returned by the user service
type UserInfo struct {
	Id                string
	Username          string
	DisplayName       string
	Email             string
	ProfilePictureUrl string
	Bio               string
	IsVerified        bool
}

// mockUserClient implements a simulated UserClient interface
type mockUserClient struct{}

// NewUserClient creates a new user client
func NewUserClient(conn interface{}) UserClient {
	if conn == nil {
		log.Println("WARNING: Using MOCK user client. Returned user data will NOT be real!")
	} else {
		log.Println("WARNING: Cannot use provided connection, using MOCK user client")
	}
	return &mockUserClient{}
}

// GetUserById simulates fetching a user by ID
func (c *mockUserClient) GetUserById(ctx context.Context, userId string) (*UserInfo, error) {
	log.Printf("WARNING: Using MOCK user data for user ID: %s", userId)

	// Generate a deterministic but varied display name
	randomGen := rand.New(rand.NewSource(int64(len(userId))))
	adjectives := []string{"Busy", "Creative", "Energetic", "Helpful", "Innovative"}
	nouns := []string{"Developer", "Designer", "Creator", "Builder", "Coder"}

	adj := adjectives[randomGen.Intn(len(adjectives))]
	noun := nouns[randomGen.Intn(len(nouns))]
	displayName := adj + " " + noun

	// Username based on display name
	username := (adj + noun)[0:6] + userId[0:4]

	// Simulate a small delay to mimic network latency
	time.Sleep(10 * time.Millisecond)

	// Return a more realistic user
	return &UserInfo{
		Id:                userId,
		Username:          username,
		DisplayName:       displayName,
		Email:             username + "@example.com",
		ProfilePictureUrl: "https://i.pravatar.cc/150?u=" + userId,
		Bio:               "This is a MOCK user profile for testing. In production, real user data would be shown.",
		IsVerified:        randomGen.Intn(2) == 1,
	}, nil
}
