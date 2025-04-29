package handlers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	authProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"
	userProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/groupcache/lru"
	supabase "github.com/supabase-community/storage-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/config"
)

var Config *config.Config

var (
	authConnPool      *ConnectionPool
	userConnPool      *ConnectionPool
	connPoolInitOnce  sync.Once
	responseCache     *lru.Cache
	requestRateLimits = make(map[string]RateLimiter)
	rateLimiterMutex  sync.RWMutex
	supabaseClient    *supabase.Client
	supabaseInitOnce  sync.Once
)

type ConnectionPool struct {
	connections chan *grpc.ClientConn
	serviceAddr string
	maxIdle     int
	maxOpen     int
	timeout     time.Duration
	mu          sync.Mutex
}

type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	tokensPerSec   float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

func NewConnectionPool(serviceAddr string, maxIdle, maxOpen int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan *grpc.ClientConn, maxIdle),
		serviceAddr: serviceAddr,
		maxIdle:     maxIdle,
		maxOpen:     maxOpen,
		timeout:     timeout,
	}
}

func InitConnectionPools() {
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		responseCache = lru.New(100)
	})
}

func (p *ConnectionPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	default:

		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()

		conn, err := grpc.DialContext(ctx, p.serviceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())

		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
	select {
	case p.connections <- conn:

	default:

		conn.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
}

func NewRateLimiter(maxTokens, tokensPerSec float64) RateLimiter {
	return RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		tokensPerSec:   tokensPerSec,
		lastRefillTime: time.Now(),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefillTime).Seconds()
	r.lastRefillTime = now

	r.tokens += elapsed * r.tokensPerSec
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}

	if r.tokens < 1 {
		return false
	}

	r.tokens--
	return true
}

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

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func init() {
	InitConnectionPools()
	supabaseInitOnce.Do(func() {
		supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
	})
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiterMutex.RLock()
		limiter, exists := requestRateLimits[ip]
		rateLimiterMutex.RUnlock()

		if !exists {
			rateLimiterMutex.Lock()

			limiter = NewRateLimiter(20, 0.33)
			requestRateLimits[ip] = limiter
			rateLimiterMutex.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Success: false,
				Message: "Rate limit exceeded. Please try again later.",
				Code:    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

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

func GetUserProfile(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &userProto.GetUserRequest{
		Id: userID,
	}

	resp, err := client.GetUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {

			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateUserProfile(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	var request userProto.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	request.Id = userID

	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.UpdateUser(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to update user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ListProducts(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "list products endpoint",
	})
}

func GetProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "get product endpoint",
	})
}

func CreateProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "create product endpoint",
	})
}

func UpdateProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "update product endpoint",
	})
}

func DeleteProduct(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "delete product endpoint",
	})
}

func CreatePayment(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "create payment endpoint",
	})
}

func GetPayment(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "get payment endpoint",
	})
}

func GetPaymentHistory(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "get payment history endpoint",
	})
}

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

func InitServices() {
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		responseCache = lru.New(100)
	})
	supabaseInitOnce.Do(func() {
		supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
	})
}
