package models

type RegisterRequest struct {
	Name                  string `json:"name" binding:"required"`
	Username              string `json:"username" binding:"required"`
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=8"`
	ConfirmPassword       string `json:"confirm_password" binding:"required,eqfield=Password"`
	Gender                string `json:"gender" binding:"required"`
	DateOfBirth           string `json:"date_of_birth" binding:"required"`
	SecurityQuestion      string `json:"securityQuestion" binding:"required"`
	SecurityAnswer        string `json:"securityAnswer" binding:"required"`
	SubscribeToNewsletter bool   `json:"subscribeToNewsletter"`
	RecaptchaToken        string `json:"recaptcha_token" binding:"required"`
	ProfilePictureUrl     string `json:"profile_picture_url,omitempty"`
	BannerUrl             string `json:"banner_url,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GoogleLoginRequest struct {
	TokenID string `json:"token_id" binding:"required"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
}

type VerifyEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}

type ResendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateUserRequest struct {
	Id                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	Username          string `json:"username,omitempty"`
	Email             string `json:"email,omitempty"`
	Gender            string `json:"gender,omitempty"`
	DateOfBirth       string `json:"date_of_birth,omitempty"`
	Bio               string `json:"bio,omitempty"`
	Location          string `json:"location,omitempty"`
	Website           string `json:"website,omitempty"`
	ProfilePictureUrl string `json:"profile_picture_url,omitempty"`
	BannerUrl         string `json:"banner_url,omitempty"`
}
