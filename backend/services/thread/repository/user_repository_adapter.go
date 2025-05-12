package repository

// UserClientAdapter adapts the UserClient to implement UserRepository
type UserClientAdapter struct {
	userClient interface {
		UserExists(userId string) (bool, error)
		GetUserDetails(userId string) (map[string]interface{}, error)
	}
}

// NewUserRepositoryAdapter creates a new adapter
func NewUserRepositoryAdapter(userClient interface{}) UserRepository {
	// Type assertion to ensure userClient implements necessary methods
	adapter := &UserClientAdapter{}

	// If userClient is not nil, check if it has the required methods
	if userClient != nil {
		if client, ok := userClient.(interface {
			UserExists(userId string) (bool, error)
			GetUserDetails(userId string) (map[string]interface{}, error)
		}); ok {
			adapter.userClient = client
		}
	}

	return adapter
}

// UserExists checks if a user exists
func (a *UserClientAdapter) UserExists(userID string) (bool, error) {
	if a.userClient == nil {
		// Default behavior if no client is available
		return true, nil // Assume user exists
	}

	return a.userClient.UserExists(userID)
}

// GetUserDetails retrieves user details
func (a *UserClientAdapter) GetUserDetails(userID string) (map[string]interface{}, error) {
	if a.userClient == nil {
		// Default behavior if no client is available
		return map[string]interface{}{}, nil
	}

	return a.userClient.GetUserDetails(userID)
}
