package repository

type UserRepository interface {
	UserExists(userID string) (bool, error)
	GetUserDetails(userID string) (map[string]interface{}, error)
}
