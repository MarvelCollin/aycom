package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Auth handler working",
		})
	}
}

func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request format")
		return
	}

	refreshSecret := utils.GetJWTSecret()
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refreshSecret, nil
	})

	if err != nil {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid refresh token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid refresh token")
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		// Try legacy format
		userID, ok = claims["user_id"].(string)
		if !ok {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token claims")
			return
		}
	}

	accessToken, err := utils.GenerateJWT(userID, time.Hour)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate new tokens")
		return
	}

	refreshTokenDuration := 7 * 24 * time.Hour
	newRefreshToken, err := utils.GenerateJWT(userID, refreshTokenDuration)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate new tokens")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    3600,
		"token_type":    "Bearer",
		"user_id":       userID,
	})
}

func GetOAuthConfig(c *gin.Context) {

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		googleClientID = "161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com"
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": []gin.H{
			{
				"name":      "google",
				"client_id": googleClientID,
				"auth_url":  "https://accounts.google.com/o/oauth2/auth",
				"scopes":    []string{"email", "profile"},
			},
		},
	})
}

func Login(c *gin.Context) {
	var req authRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if userServiceClient == nil {
		log.Println("Login: User service client is nil, attempting to initialize")
		InitUserServiceClient(AppConfig)
		if userServiceClient == nil {
			log.Println("Login: Failed to initialize user service client")
			utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
			return
		}
		log.Println("Login: User service client initialized successfully")
	}

	userAuthResp, err := userServiceClient.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("Login: Failed to authenticate user: %v", err)
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Get the User object from the response
	user := userAuthResp.User
	if user == nil {
		log.Printf("Login: User is nil in authentication response")
		utils.SendErrorResponse(c, http.StatusInternalServerError, "AUTH_ERROR", "User data not available")
		return
	}

	log.Printf("Login: User authenticated successfully: %s (ID: %s)", user.Email, user.ID)

	// Add more user claims to the token
	tokenClaims := jwt.MapClaims{
		"sub":         user.ID,
		"user_id":     user.ID, // For backward compatibility
		"email":       user.Email,
		"username":    user.Username,
		"is_admin":    user.IsAdmin,
		"is_verified": user.IsVerified,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	log.Printf("Login: Generated token claims for %s with admin status: %t", user.Email, user.IsAdmin)

	// For debugging, log all claims
	for k, v := range tokenClaims {
		log.Printf("Login: Token claim %s: %v (type: %T)", k, v, v)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString(utils.GetJWTSecret())
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub":        user.ID,
		"user_id":    user.ID, // For backward compatibility
		"token_type": "refresh",
		"is_admin":   user.IsAdmin,
		"exp":        time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(utils.GetJWTSecret())
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":                  user.ID,
			"name":                user.Name,
			"username":            user.Username,
			"email":               user.Email,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"created_at":          user.CreatedAt,
		},
		"token":         tokenString,
		"refresh_token": refreshTokenString,
	})
}

func VerifyEmail(c *gin.Context) {
	var req struct {
		Email            string `json:"email" binding:"required,email"`
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request format. Email and verification code are required.")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil {
		log.Printf("Error fetching user by email: %v", err)
		utils.SendErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	if userResp.User == nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	updateReq := &userProto.UpdateUserVerificationStatusRequest{
		UserId:     userResp.User.Id,
		IsVerified: true,
	}

	updateResp, err := UserClient.UpdateUserVerificationStatus(ctx, updateReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Invalid verification code")
			case codes.NotFound:
				utils.SendErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "VERIFICATION_FAILED", "Failed to verify email: "+st.Message())
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "VERIFICATION_FAILED", "Failed to verify email")
		}
		return
	}

	if updateResp.Success {
		accessToken, err := utils.GenerateJWT(userResp.User.Id, time.Hour)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Email verified but failed to generate token")
			return
		}

		refreshTokenDuration := 7 * 24 * time.Hour
		refreshToken, err := utils.GenerateJWT(userResp.User.Id, refreshTokenDuration)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Email verified but failed to generate token")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"message":       "Email verified successfully",
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user_id":       userResp.User.Id,
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	} else {
		utils.SendErrorResponse(c, http.StatusBadRequest, "VERIFICATION_FAILED", updateResp.Message)
	}
}

func ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid email format")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "If the email exists, a new verification code has been sent",
			})
			return
		}

		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to process request")
		return
	}

	if userResp.User == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "If the email exists, a new verification code has been sent",
		})
		return
	}

	if userResp.User.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User is already verified",
		})
		return
	}

	code := utils.GenerateVerificationCode()

	log.Printf("New verification code for %s: %s", req.Email, code)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification code has been sent to your email",
	})
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func GoogleLogin(c *gin.Context) {
	var req struct {
		TokenID string `json:"token_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid token format")
		return
	}

	googleAPIURL := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + req.TokenID
	resp, err := http.Get(googleAPIURL)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "GOOGLE_API_ERROR", "Failed to verify Google token")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid Google token")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "READ_ERROR", "Failed to read Google API response")
		return
	}

	var tokenInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "PARSE_ERROR", "Failed to parse Google API response")
		return
	}

	// Check if user service is available, and try to reinitialize if needed
	if UserClient == nil {
		log.Println("GoogleLogin: User service client is nil, attempting to initialize")
		InitGRPCServices()
		InitUserServiceClient(AppConfig)

		if UserClient == nil {
			log.Println("GoogleLogin: Failed to initialize user service client")
			utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
			return
		}
		log.Println("GoogleLogin: User service client initialized successfully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userResp *userProto.GetUserByEmailResponse
	var getUserErr error

	// Add retry logic for getting user by email
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("GoogleLogin: Retry attempt %d/%d for GetUserByEmail", i+1, maxRetries)
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}

		userResp, getUserErr = UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
			Email: tokenInfo.Email,
		})

		if getUserErr == nil {
			break
		}

		log.Printf("GoogleLogin: Error getting user by email (attempt %d): %v", i+1, getUserErr)
	}

	var userID string
	var isNewUser bool
	var requiresProfileCompletion bool

	if getUserErr != nil || userResp == nil || userResp.User == nil {
		log.Printf("GoogleLogin: User with email %s not found, creating new user", tokenInfo.Email)

		newUsername := utils.GenerateUsername(tokenInfo.Name)
		randomPassword := utils.GenerateSecureRandomPassword(16)
		currentDate := time.Now().Format("2006-01-02")

		user := &userProto.User{
			Name:              tokenInfo.Name,
			Username:          newUsername,
			Email:             tokenInfo.Email,
			Password:          randomPassword,
			ProfilePictureUrl: tokenInfo.Picture,
			DateOfBirth:       currentDate,
			Gender:            "unknown",
			IsVerified:        true,
		}

		createReq := &userProto.CreateUserRequest{
			User: user,
		}

		// Add retry logic for creating a new user
		var createResp *userProto.CreateUserResponse
		var createErr error

		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				log.Printf("GoogleLogin: Retry attempt %d/%d for CreateUser", i+1, maxRetries)
				time.Sleep(time.Duration(i) * 500 * time.Millisecond)
			}

			createResp, createErr = UserClient.CreateUser(ctx, createReq)

			if createErr == nil && createResp != nil && createResp.User != nil {
				break
			}

			log.Printf("GoogleLogin: Error creating user (attempt %d): %v", i+1, createErr)
		}

		if createErr != nil || createResp == nil || createResp.User == nil {
			log.Printf("GoogleLogin: Failed to create user account: %v", createErr)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "USER_CREATION_FAILED", "Failed to create user account")
			return
		}

		userID = createResp.User.Id
		isNewUser = true
		requiresProfileCompletion = true
		log.Printf("GoogleLogin: Successfully created new user with ID: %s", userID)
	} else {
		userID = userResp.User.Id
		log.Printf("GoogleLogin: Found existing user with ID: %s", userID)

		// Check if user needs to complete their profile
		user := userResp.User
		if user.Gender == "" || user.Gender == "unknown" ||
			user.DateOfBirth == "" ||
			user.SecurityQuestion == "" ||
			user.SecurityAnswer == "" {
			requiresProfileCompletion = true
			log.Printf("GoogleLogin: User %s needs to complete their profile", userID)
		}
	}

	accessToken, err := utils.GenerateJWT(userID, time.Hour)
	if err != nil {
		log.Printf("GoogleLogin: Failed to generate access token: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	refreshTokenDuration := 7 * 24 * time.Hour
	refreshToken, err := utils.GenerateJWT(userID, refreshTokenDuration)
	if err != nil {
		log.Printf("GoogleLogin: Failed to generate refresh token: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	log.Printf("GoogleLogin: Authentication successful for user ID: %s", userID)
	c.JSON(http.StatusOK, gin.H{
		"success":                     true,
		"message":                     "Google authentication successful",
		"access_token":                accessToken,
		"refresh_token":               refreshToken,
		"user_id":                     userID,
		"expires_in":                  3600,
		"token_type":                  "Bearer",
		"is_new_user":                 isNewUser,
		"requires_profile_completion": requiresProfileCompletion,
	})
}

func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid email format")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil || userResp.User == nil {
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"message": "If the email address exists in our system, a security question will be sent",
		})
		return
	}

	securityQuestion := userResp.User.SecurityQuestion
	if securityQuestion == "" {
		securityQuestion = "What is your mother's maiden name?"
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":           "Please answer your security question to reset your password",
		"security_question": securityQuestion,
		"email":             req.Email,
	})
}

func VerifySecurityAnswer(c *gin.Context) {
	var req struct {
		Email          string `json:"email" binding:"required,email"`
		SecurityAnswer string `json:"security_answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request format")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil || userResp.User == nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", "User not found")
		return
	}

	if userResp.User.SecurityAnswer != req.SecurityAnswer {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_ANSWER", "Security answer is incorrect")
		return
	}

	resetToken := utils.GetTokenManager().Generate(req.Email)

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Security answer verified. You may now reset your password.",
		"token":   resetToken.Token,
		"email":   req.Email,
		"expires": resetToken.ExpiresAt,
	})
}

func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request. Token, email, and password (min 8 chars) are required.")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	if !utils.GetTokenManager().Validate(req.Token, req.Email) {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_TOKEN", "Reset token is invalid or expired")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: req.Email,
	})

	if err != nil || userResp.User == nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", "User not found")
		return
	}

	// We can't compare the old password hash with the new plain text password
	// So we'll just update the password without checking if it's the same

	updateReq := &userProto.UpdateUserRequest{
		UserId: userResp.User.Id,
		User: &userProto.User{
			Id:       userResp.User.Id,
			Password: req.NewPassword,
		},
	}

	_, err = UserClient.UpdateUser(ctx, updateReq)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "PASSWORD_UPDATE_FAILED", "Failed to update password")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Password has been reset successfully. You can now log in with your new password",
	})

	utils.GetTokenManager().Delete(req.Token)
}

// CheckAdminStatus checks if the authenticated user has admin privileges
func CheckAdminStatus(c *gin.Context) {
	log.Printf("CheckAdminStatus: Processing admin status check request")
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("CheckAdminStatus: No userID in context")
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)
	log.Printf("CheckAdminStatus: Processing request for user %s", userIDStr)

	// First try to get from JWT claims
	tokenString := c.GetHeader("Authorization")
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:] // Remove Bearer prefix
		log.Printf("CheckAdminStatus: Token length: %d", len(tokenString))

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(utils.GetJWTSecret()), nil
		})

		if err == nil && token.Valid {
			claims, ok := token.Claims.(jwt.MapClaims)
			if ok {
				log.Printf("CheckAdminStatus: Valid JWT token parsed")

				// Log all claims for debugging
				log.Printf("CheckAdminStatus: All JWT claims:")
				for k, v := range claims {
					log.Printf("  %s: %v (Type: %T)", k, v, v)
				}

				// Try to get admin status from various claim formats
				isAdmin := false

				// Check common claim names for is_admin
				if adminValue, exists := claims["is_admin"]; exists {
					log.Printf("CheckAdminStatus: Found is_admin claim: %v (Type: %T)", adminValue, adminValue)
					switch v := adminValue.(type) {
					case bool:
						isAdmin = v
						log.Printf("CheckAdminStatus: is_admin as bool: %t", isAdmin)
					case string:
						isAdmin = v == "true" || v == "t" || v == "1"
						log.Printf("CheckAdminStatus: is_admin as string: %s -> %t", v, isAdmin)
					case float64: // JSON numbers are parsed as float64
						isAdmin = v == 1
						log.Printf("CheckAdminStatus: is_admin as float64: %f -> %t", v, isAdmin)
					}
				}

				// Also check for other possible formats
				if !isAdmin {
					if adminValue, exists := claims["admin"]; exists {
						log.Printf("CheckAdminStatus: Found admin claim: %v (Type: %T)", adminValue, adminValue)
						switch v := adminValue.(type) {
						case bool:
							isAdmin = v
							log.Printf("CheckAdminStatus: admin as bool: %t", isAdmin)
						case string:
							isAdmin = v == "true" || v == "t" || v == "1"
							log.Printf("CheckAdminStatus: admin as string: %s -> %t", v, isAdmin)
						case float64:
							isAdmin = v == 1
							log.Printf("CheckAdminStatus: admin as float64: %f -> %t", v, isAdmin)
						}
					}
				}

				// If admin status confirmed by JWT, return immediately
				if isAdmin {
					log.Printf("CheckAdminStatus: User %s is admin according to JWT claims", userIDStr)
					utils.SendSuccessResponse(c, http.StatusOK, gin.H{
						"is_admin": true,
						"user_id":  userIDStr,
						"source":   "jwt",
					})
					return
				}

				log.Printf("CheckAdminStatus: User %s is not admin according to JWT claims, checking database", userIDStr)
			}
		} else {
			log.Printf("CheckAdminStatus: Invalid token or parsing error: %v", err)
		}
	} else {
		log.Printf("CheckAdminStatus: No valid Bearer token found")
	}

	// If we get here, either the token didn't have the is_admin claim or we couldn't parse it
	// So we'll need to check the database
	if userServiceClient == nil {
		log.Printf("CheckAdminStatus: User service unavailable")
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	log.Printf("CheckAdminStatus: Checking database for user %s", userIDStr)
	user, err := userServiceClient.GetUserById(userIDStr)
	if err != nil {
		log.Printf("CheckAdminStatus: Failed to get user by ID %s: %v", userIDStr, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check admin status")
		return
	}

	log.Printf("CheckAdminStatus: User %s has admin status from database: %t", userIDStr, user.IsAdmin)

	// Generate a new token with admin claim if user is admin
	if user.IsAdmin {
		log.Printf("CheckAdminStatus: User %s is admin, generating new token with admin claim", userIDStr)
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"is_admin": user.IsAdmin,
		"user_id":  userIDStr,
		"source":   "database",
	})
}
