package handlers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	authProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	supabase "github.com/supabase-community/storage-go"
	"google.golang.org/grpc/status"
)

// Auth request/response types
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
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

// Login handles user authentication
func Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Login failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// Register creates a new user account
func Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.RegisterRequest{
		Name:                  request.Name,
		Username:              request.Username,
		Email:                 request.Email,
		Password:              request.Password,
		ConfirmPassword:       request.ConfirmPassword,
		Gender:                request.Gender,
		DateOfBirth:           request.DateOfBirth,
		SecurityQuestion:      request.SecurityQuestion,
		SecurityAnswer:        request.SecurityAnswer,
		SubscribeToNewsletter: request.SubscribeToNewsletter,
		RecaptchaToken:        request.RecaptchaToken,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Registration failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

// RefreshToken issues a new access token using a refresh token
func RefreshToken(c *gin.Context) {
	var request RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
	}

	resp, err := client.RefreshToken(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Token refresh failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// GoogleAuth authenticates a user using Google OAuth
func GoogleAuth(c *gin.Context) {
	var requestBody struct {
		TokenID string `json:"token_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.GoogleLoginRequest{
		IdToken: requestBody.TokenID,
	}

	resp, err := client.GoogleLogin(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Google authentication failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// VerifyEmail verifies a user's email using a verification code
func VerifyEmail(c *gin.Context) {
	var request VerifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.VerifyEmailRequest{
		Email:            request.Email,
		VerificationCode: request.VerificationCode,
	}

	resp, err := client.VerifyEmail(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Email verification failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// ResendVerificationCode resends a verification code to the user's email
func ResendVerificationCode(c *gin.Context) {
	var request ResendCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.ResendVerificationCodeRequest{
		Email: request.Email,
	}

	resp, err := client.ResendVerificationCode(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to resend verification code: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
	})
}

// Helper function to upload files to Supabase storage
func uploadToSupabase(fileHeader *multipart.FileHeader, bucketName string, destinationPath string) (string, error) {
	if fileHeader == nil {
		return "", nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	upsert := false
	fileOptions := supabase.FileOptions{
		ContentType: &contentType,
		Upsert:      &upsert,
	}

	_, err = supabaseClient.UploadFile(bucketName, destinationPath, file, fileOptions)

	if err != nil {
		return "", fmt.Errorf("failed to upload to supabase: %w", err)
	}

	publicURL := supabaseClient.GetPublicUrl(bucketName, destinationPath)

	return publicURL.SignedURL, nil
}

// RegisterWithMedia handles registration with profile images
func RegisterWithMedia(c *gin.Context) {
	if supabaseClient == nil {
		InitServices()
	}
	if supabaseClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Supabase client not initialized", Code: "CONFIG_ERROR"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid form data: " + err.Error(), Code: "INVALID_REQUEST"})
		return
	}

	values := form.Value
	name := values["name"][0]
	username := values["username"][0]
	email := values["email"][0]
	password := values["password"][0]
	confirmPassword := values["confirm_password"][0]
	gender := values["gender"][0]
	dateOfBirth := values["date_of_birth"][0]
	securityQuestion := values["security_question"][0]
	securityAnswer := values["security_answer"][0]
	subscribe := values["subscribe_to_newsletter"][0] == "true"
	recaptchaToken := values["recaptcha_token"][0]

	if password != confirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Passwords do not match", Code: "VALIDATION_ERROR"})
		return
	}

	profilePicFileHeader, _ := c.FormFile("profile_picture")
	bannerFileHeader, _ := c.FormFile("banner_image")

	uuidVal, _ := uuid.NewV4()
	userPathPrefix := uuidVal.String()

	profilePicURL, err := uploadToSupabase(profilePicFileHeader, "profile-pictures", userPathPrefix+"/"+filepath.Base(profilePicFileHeader.Filename))
	if err != nil {
		log.Printf("Failed to upload profile picture: %v", err)
	}
	bannerURL, err := uploadToSupabase(bannerFileHeader, "banner-images", userPathPrefix+"/"+filepath.Base(bannerFileHeader.Filename))
	if err != nil {
		log.Printf("Failed to upload banner image: %v", err)
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to connect to auth service: " + err.Error(), Code: "SERVICE_UNAVAILABLE"})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	req := &authProto.RegisterRequest{
		Name:                  name,
		Username:              username,
		Email:                 email,
		Password:              password,
		ConfirmPassword:       confirmPassword,
		Gender:                gender,
		DateOfBirth:           dateOfBirth,
		SecurityQuestion:      securityQuestion,
		SecurityAnswer:        securityAnswer,
		SubscribeToNewsletter: subscribe,
		RecaptchaToken:        recaptchaToken,
		ProfilePictureUrl:     profilePicURL,
		BannerUrl:             bannerURL,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: st.Message(), Code: st.Code().String()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Registration failed: " + err.Error(), Code: "INTERNAL_ERROR"})
		}
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}
