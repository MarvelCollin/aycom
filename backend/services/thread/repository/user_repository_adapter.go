package repository

type UserClientAdapter struct {
	userClient interface {
		UserExists(userId string) (bool, error)
		GetUserDetails(userId string) (map[string]interface{}, error)
	}
}

func NewUserRepositoryAdapter(userClient interface{}) UserRepository {

	adapter := &UserClientAdapter{}

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

func (a *UserClientAdapter) UserExists(userID string) (bool, error) {
	if a.userClient == nil {

		return true, nil 
	}

	return a.userClient.UserExists(userID)
}

func (a *UserClientAdapter) GetUserDetails(userID string) (map[string]interface{}, error) {
	if a.userClient == nil {

		return map[string]interface{}{}, nil
	}

	return a.userClient.GetUserDetails(userID)
}