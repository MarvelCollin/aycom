package handlers

import (
	"net/http"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/config"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	authProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"
)

// Global config for the handlers
var Config *config.Config

// RegisterRequest represents the user registration payload
type RegisterRequest struct {
	Name                  string `json:"name" binding:"required"`
	Username              string `json:"username" binding:"required"`
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=8"`
	ConfirmPassword       string `json:"confirmPassword" binding:"required,eqfield=Password"`
	Gender                string `json:"gender" binding:"required"`
	DateOfBirth           string `json:"dateOfBirth" binding:"required"`
	SecurityQuestion      string `json:"securityQuestion" binding:"required"`
	SecurityAnswer        string `json:"securityAnswer" binding:"required"`
	SubscribeToNewsletter bool   `json:"subscribeToNewsletter"`
	RecaptchaToken        string `json:"recaptcha_token" binding:"required"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerifyEmailRequest represents the request for verifying email
type VerifyEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}

// ResendCodeRequest represents the request for resending verification code
type ResendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// LogoutRequest represents the request for logging out
type LogoutRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenRequest represents the request for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterResponse represents the response from user registration
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetOAuthConfig godoc
// @Summary Get OAuth configuration
// @Description Get OAuth configuration details for the client
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "OAuth configuration"
// @Router /api/v1/auth/oauth-config [get]
func GetOAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"google_client_id": Config.OAuth.GoogleClientID,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

	req := &authProto.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	resp, err := client.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Login failed: " + err.Error(),
		})
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

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body RegisterRequest true "User registration data"
// @Success 200 {object} RegisterResponse "registration successful"
// @Failure 400 {object} ErrorResponse "invalid input"
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

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

	resp, err := client.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Registration failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

func RefreshToken(c *gin.Context) {
	var request RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

	req := &authProto.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
	}

	resp, err := client.RefreshToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Token refresh failed: " + err.Error(),
		})
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

// GoogleAuth godoc
// @Summary Authenticate with Google
// @Description Use Google OAuth token to authenticate
// @Tags auth
// @Accept json
// @Produce json
// @Param token body map[string]interface{} true "Google token ID"
// @Success 200 {object} map[string]interface{} "tokens and user info"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Router /api/v1/auth/google [post]
func GoogleAuth(c *gin.Context) {
	var requestBody struct {
		TokenID string `json:"token_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

	req := &authProto.GoogleLoginRequest{
		TokenId: requestBody.TokenID,
	}

	resp, err := client.GoogleLogin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Google authentication failed: " + err.Error(),
		})
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

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user's email address with verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param verification body map[string]interface{} true "Email and verification code"
// @Success 200 {object} map[string]interface{} "tokens and user info"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/auth/verify-email [post]
func VerifyEmail(c *gin.Context) {
	var request VerifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

	req := &authProto.VerifyEmailRequest{
		Email:            request.Email,
		VerificationCode: request.VerificationCode,
	}

	resp, err := client.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Email verification failed: " + err.Error(),
		})
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

// ResendVerificationCode godoc
// @Summary Resend verification code
// @Description Resend verification code to user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param email body map[string]interface{} true "User email"
// @Success 200 {object} map[string]interface{} "success message"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/auth/resend-code [post]
func ResendVerificationCode(c *gin.Context) {
	var request ResendCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := grpc.Dial(Config.GetAuthServiceAddr(), grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	client := authProto.NewAuthServiceClient(conn)

	req := &authProto.ResendVerificationCodeRequest{
		Email: request.Email,
	}

	resp, err := client.ResendVerificationCode(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to resend verification code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "user profile"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Router /api/v1/users/profile [get]
func GetUserProfile(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the user service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get user profile endpoint",
	})
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body map[string]interface{} true "User profile data"
// @Success 200 {object} map[string]interface{} "updated profile"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Router /api/v1/users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the user service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "update user profile endpoint",
	})
}

// ListProducts godoc
// @Summary List products
// @Description Get a list of products
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "list of products"
// @Router /api/v1/products [get]
func ListProducts(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "list products endpoint",
	})
}

// GetProduct godoc
// @Summary Get product
// @Description Get a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "product details"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [get]
func GetProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get product endpoint",
	})
}

// CreateProduct godoc
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body map[string]interface{} true "Product data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "created product"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/products [post]
func CreateProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "create product endpoint",
	})
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body map[string]interface{} true "Updated product data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "updated product"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [put]
func UpdateProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "update product endpoint",
	})
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 204 "no content"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "delete product endpoint",
	})
}

// CreatePayment godoc
// @Summary Create payment
// @Description Create a new payment
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body map[string]interface{} true "Payment data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "created payment"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/payments [post]
func CreatePayment(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "create payment endpoint",
	})
}

// GetPayment godoc
// @Summary Get payment
// @Description Get a payment by ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "payment details"
// @Failure 404 {object} map[string]interface{} "payment not found"
// @Router /api/v1/payments/{id} [get]
func GetPayment(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment endpoint",
	})
}

// GetPaymentHistory godoc
// @Summary Get payment history
// @Description Get payment history for the authenticated user
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "payment history"
// @Router /api/v1/payments/history [get]
func GetPaymentHistory(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment history endpoint",
	})
}
