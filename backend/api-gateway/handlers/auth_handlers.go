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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

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
		utils.SendErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token claims")
		return
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
	c.JSON(http.StatusOK, gin.H{
		"message": "Login endpoint. Please use /api/v1/users/login instead.",
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

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUserByEmail(ctx, &userProto.GetUserByEmailRequest{
		Email: tokenInfo.Email,
	})

	var userID string
	if err != nil || userResp.User == nil {

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

		createResp, err := UserClient.CreateUser(ctx, createReq)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "USER_CREATION_FAILED", "Failed to create user account")
			return
		}

		userID = createResp.User.Id
	} else {

		userID = userResp.User.Id
	}

	accessToken, err := utils.GenerateJWT(userID, time.Hour)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	refreshTokenDuration := 7 * 24 * time.Hour
	refreshToken, err := utils.GenerateJWT(userID, refreshTokenDuration)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Google authentication successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       userID,
		"expires_in":    3600,
		"token_type":    "Bearer",
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

	storedPassword := userResp.User.Password
	if storedPassword == req.NewPassword {
		utils.SendErrorResponse(c, http.StatusBadRequest, "SAME_PASSWORD", "New password cannot be the same as the old one")
		return
	}

	updateReq := &userProto.UpdateUserRequest{
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
