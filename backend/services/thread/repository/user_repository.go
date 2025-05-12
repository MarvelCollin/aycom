package repository

// UserRepository defines the minimum interface for user-related operations
// needed by thread services
type UserRepository interface {
	// Basic user operations that might be needed
	UserExists(userID string) (bool, error)
	GetUserDetails(userID string) (map[string]interface{}, error)
}
