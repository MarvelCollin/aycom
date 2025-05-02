package dto

type SessionTokenDTO struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	ExpiresAt    string `json:"expires_at"`
	CreatedAt    string `json:"created_at"`
}
