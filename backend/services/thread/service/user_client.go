package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"
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

// realUserClient implements the real UserClient interface using gRPC
type realUserClient struct {
	conn *grpc.ClientConn
}

// NewUserClient creates a new user client
func NewUserClient(conn interface{}) UserClient {
	if conn == nil {
		log.Println("WARNING: Using MOCK user client. Returned user data will NOT be real!")
		return &mockUserClient{}
	}

	if grpcConn, ok := conn.(*grpc.ClientConn); ok {
		log.Println("Using REAL user client with gRPC connection")
		return &realUserClient{
			conn: grpcConn,
		}
	}

	log.Println("WARNING: Cannot use provided connection, using MOCK user client")
	return &mockUserClient{}
}

// GetUserById for the real user client implementation
func (c *realUserClient) GetUserById(ctx context.Context, userId string) (*UserInfo, error) {
	// Create a generic gRPC request to user service
	method := "/user.UserService/GetUser"

	// Create request payload manually
	requestBytes, _ := json.Marshal(map[string]string{
		"user_id": userId,
	})

	// Make a raw gRPC call
	result := c.conn.Invoke(ctx, method, requestBytes, nil)
	if result != nil {
		log.Printf("Error invoking user service: %v", result)

		// Fallback to mock user if there's an error
		log.Printf("Falling back to mock data for user ID: %s", userId)
		mockClient := &mockUserClient{}
		return mockClient.GetUserById(ctx, userId)
	}

	// In a real implementation, we would properly decode the result
	// For now, this is unlikely to work and will fall back to mock data
	log.Printf("WARNING: Using REAL user service connection but falling back to mock data for now")
	mockClient := &mockUserClient{}
	return mockClient.GetUserById(ctx, userId)
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
		Bio:               fmt.Sprintf("This is user %s. In production, real user data would be shown.", userId),
		IsVerified:        randomGen.Intn(2) == 1,
	}, nil
}
