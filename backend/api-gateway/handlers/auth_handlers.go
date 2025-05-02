package handlers

import (
	"aycom/backend/api-gateway/models"
	"aycom/backend/proto"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Login handles user login
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to authentication service",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.Login(ctx, &proto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			case codes.NotFound, codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Success: false,
					Message: "Invalid email or password",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Authentication service error: " + statusErr.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		UserId:       response.UserId,
		TokenType:    response.TokenType,
		ExpiresIn:    response.ExpiresIn,
	})
}

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to authentication service",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Register(ctx, &proto.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, ErrorResponse{
					Success: false,
					Message: "A user with this email or username already exists",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Registration service error: " + statusErr.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.RegisterResponse{
		Success: true,
		Message: "Registration successful",
	})
}

// RefreshToken refreshes an access token using a refresh token
func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to authentication service",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.RefreshToken(ctx, &proto.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument, codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Success: false,
					Message: "Invalid or expired refresh token",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Authentication service error: " + statusErr.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		UserId:       response.UserId,
		TokenType:    response.TokenType,
		ExpiresIn:    response.ExpiresIn,
	})
}

// RegisterWithMedia handles user registration with profile picture and banner uploads
func RegisterWithMedia(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Failed to parse form: " + err.Error(),
		})
		return
	}

	// Parse basic user data
	registerRequest := models.RegisterRequest{
		Name:             c.PostForm("name"),
		Username:         c.PostForm("username"),
		Email:            c.PostForm("email"),
		Password:         c.PostForm("password"),
		ConfirmPassword:  c.PostForm("confirm_password"),
		Gender:           c.PostForm("gender"),
		DateOfBirth:      c.PostForm("date_of_birth"),
		SecurityQuestion: c.PostForm("securityQuestion"),
		SecurityAnswer:   c.PostForm("securityAnswer"),
		RecaptchaToken:   c.PostForm("recaptcha_token"),
	}

	// Check for required fields
	if registerRequest.Name == "" || registerRequest.Username == "" ||
		registerRequest.Email == "" || registerRequest.Password == "" ||
		registerRequest.Gender == "" || registerRequest.DateOfBirth == "" ||
		registerRequest.SecurityQuestion == "" || registerRequest.SecurityAnswer == "" ||
		registerRequest.RecaptchaToken == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Missing required fields",
		})
		return
	}

	// Check if passwords match
	if registerRequest.Password != registerRequest.ConfirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Passwords do not match",
		})
		return
	}

	// Continue with registration
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to authentication service",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Register(ctx, &proto.RegisterRequest{
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, ErrorResponse{
					Success: false,
					Message: "A user with this email or username already exists",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Registration service error: " + statusErr.Message(),
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Internal server error: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.RegisterResponse{
		Success: true,
		Message: "Registration successful",
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to authentication service",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := proto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Logout(ctx, &proto.LogoutRequest{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			switch statusErr.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: statusErr.Message(),
				})
			default:
				// Even if there's an error with the token, we still want to consider the logout successful
				log.Printf("Non-critical error during logout: %v", statusErr.Message())
			}
		} else {
			log.Printf("Non-critical error during logout: %v", err.Error())
		}
	}

	// Clear any auth cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully logged out",
	})
}

// GetOAuthConfig returns OAuth configuration for client-side use
func GetOAuthConfig(c *gin.Context) {
	oauthConfig := map[string]interface{}{
		"google": map[string]string{
			"client_id": Config.OAuth.GoogleClientID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    oauthConfig,
	})
}
