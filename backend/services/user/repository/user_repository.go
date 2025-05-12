// UserRepository defines methods for user data access
type UserRepository interface {
	// ... existing methods ...

	// Check if user exists by ID
	UserExists(userID string) (bool, error)

	// ... existing methods ...
}

// ... existing code ...

// UserExists checks if a user exists by ID
func (r *PostgresUserRepository) UserExists(userID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID format: %w", err)
	}

	var count int64
	result := r.db.Model(&model.User{}).
		Where("id = ?", userUUID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("error checking user existence: %w", result.Error)
	}

	return count > 0, nil
} 