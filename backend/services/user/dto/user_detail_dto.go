package dto

type UserDetailDTO struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Username               string `json:"username"`
	Email                  string `json:"email"`
	Gender                 string `json:"gender"`
	DateOfBirth            string `json:"date_of_birth"`
	ProfilePictureURL      string `json:"profile_picture_url"`
	BannerURL              string `json:"banner_url"`
	IsActivated            bool   `json:"is_activated"`
	IsBanned               bool   `json:"is_banned"`
	IsDeactivated          bool   `json:"is_deactivated"`
	IsAdmin                bool   `json:"is_admin"`
	NewsletterSubscription bool   `json:"newsletter_subscription"`
	LastLoginAt            string `json:"last_login_at"`
	JoinedAt               string `json:"joined_at"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}
