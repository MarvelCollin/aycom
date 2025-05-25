package handlers

import (
	userProto "aycom/backend/proto/user"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BanUser handles banning/unbanning users
func BanUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	var req struct {
		Ban    bool   `json:"ban" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("BanUser Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.BanUser(ctx, &userProto.BanUserRequest{
		UserId: userID,
		Ban:    req.Ban,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("BanUser Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user ban status")
			}
		} else {
			log.Printf("BanUser Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user ban status")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// SendNewsletter handles sending newsletters to subscribed users
func SendNewsletter(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("SendNewsletter Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("userID")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin not authenticated")
		return
	}
	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := UserClient.SendNewsletter(ctx, &userProto.SendNewsletterRequest{
		Subject: req.Subject,
		Content: req.Content,
		AdminId: adminID.(string),
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("SendNewsletter Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send newsletter")
			}
		} else {
			log.Printf("SendNewsletter Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send newsletter")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          response.Success,
		"message":          response.Message,
		"recipients_count": response.RecipientsCount,
	})
}

// GetCommunityRequests handles getting community creation requests
func GetCommunityRequests(c *gin.Context) {
	page := 1
	limit := 10
	status := ""

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status = statusStr
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.GetCommunityRequests(ctx, &userProto.GetCommunityRequestsRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Status: status,
	})

	if err != nil {
		log.Printf("GetCommunityRequests Handler: gRPC error: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get community requests")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"requests":    response.Requests,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
	})
}

// ProcessCommunityRequest handles approving/rejecting community creation requests
func ProcessCommunityRequest(c *gin.Context) {
	requestID := c.Param("requestId")
	if requestID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessCommunityRequest Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.ProcessCommunityRequest(ctx, &userProto.ProcessCommunityRequestRequest{
		RequestId: requestID,
		Approve:   req.Approve,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community request not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessCommunityRequest Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process community request")
			}
		} else {
			log.Printf("ProcessCommunityRequest Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process community request")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// GetPremiumRequests handles getting premium user requests
func GetPremiumRequests(c *gin.Context) {
	page := 1
	limit := 10
	status := ""

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status = statusStr
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.GetPremiumRequests(ctx, &userProto.GetPremiumRequestsRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Status: status,
	})

	if err != nil {
		log.Printf("GetPremiumRequests Handler: gRPC error: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get premium requests")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"requests":    response.Requests,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
	})
}

// ProcessPremiumRequest handles approving/rejecting premium user requests
func ProcessPremiumRequest(c *gin.Context) {
	requestID := c.Param("requestId")
	if requestID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessPremiumRequest Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.ProcessPremiumRequest(ctx, &userProto.ProcessPremiumRequestRequest{
		RequestId: requestID,
		Approve:   req.Approve,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Premium request not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessPremiumRequest Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process premium request")
			}
		} else {
			log.Printf("ProcessPremiumRequest Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process premium request")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// GetReportRequests handles getting user report requests
func GetReportRequests(c *gin.Context) {
	page := 1
	limit := 10
	status := ""

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status = statusStr
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.GetReportRequests(ctx, &userProto.GetReportRequestsRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Status: status,
	})

	if err != nil {
		log.Printf("GetReportRequests Handler: gRPC error: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get report requests")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"requests":    response.Requests,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
	})
}

// ProcessReportRequest handles approving/rejecting user report requests
func ProcessReportRequest(c *gin.Context) {
	requestID := c.Param("requestId")
	if requestID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessReportRequest Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.ProcessReportRequest(ctx, &userProto.ProcessReportRequestRequest{
		RequestId: requestID,
		Approve:   req.Approve,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Report request not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessReportRequest Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process report request")
			}
		} else {
			log.Printf("ProcessReportRequest Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process report request")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// GetThreadCategories handles getting thread categories
func GetThreadCategories(c *gin.Context) {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.GetThreadCategories(ctx, &userProto.GetThreadCategoriesRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})

	if err != nil {
		log.Printf("GetThreadCategories Handler: gRPC error: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get thread categories")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"categories":  response.Categories,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
	})
}

// CreateThreadCategory handles creating thread categories
func CreateThreadCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateThreadCategory Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.CreateThreadCategory(ctx, &userProto.CreateThreadCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("CreateThreadCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create thread category")
			}
		} else {
			log.Printf("CreateThreadCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create thread category")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

// UpdateThreadCategory handles updating thread categories
func UpdateThreadCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateThreadCategory Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.UpdateThreadCategory(ctx, &userProto.UpdateThreadCategoryRequest{
		Id:          categoryID,
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread category not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("UpdateThreadCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread category")
			}
		} else {
			log.Printf("UpdateThreadCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

// DeleteThreadCategory handles deleting thread categories
func DeleteThreadCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.DeleteThreadCategory(ctx, &userProto.DeleteThreadCategoryRequest{
		Id: categoryID,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread category not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("DeleteThreadCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete thread category")
			}
		} else {
			log.Printf("DeleteThreadCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete thread category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

// GetCommunityCategories handles getting community categories
func GetCommunityCategories(c *gin.Context) {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.GetCommunityCategories(ctx, &userProto.GetCommunityCategoriesRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})

	if err != nil {
		log.Printf("GetCommunityCategories Handler: gRPC error: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get community categories")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"categories":  response.Categories,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
	})
}

// CreateCommunityCategory handles creating community categories
func CreateCommunityCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateCommunityCategory Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.CreateCommunityCategory(ctx, &userProto.CreateCommunityCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("CreateCommunityCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create community category")
			}
		} else {
			log.Printf("CreateCommunityCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create community category")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

// UpdateCommunityCategory handles updating community categories
func UpdateCommunityCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateCommunityCategory Handler: Invalid request payload: %v", err)
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.UpdateCommunityCategory(ctx, &userProto.UpdateCommunityCategoryRequest{
		Id:          categoryID,
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community category not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("UpdateCommunityCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update community category")
			}
		} else {
			log.Printf("UpdateCommunityCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update community category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

// DeleteCommunityCategory handles deleting community categories
func DeleteCommunityCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	if UserClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := UserClient.DeleteCommunityCategory(ctx, &userProto.DeleteCommunityCategoryRequest{
		Id: categoryID,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community category not found")
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("DeleteCommunityCategory Handler: gRPC error: %v", err)
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete community category")
			}
		} else {
			log.Printf("DeleteCommunityCategory Handler: Unknown error: %v", err)
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete community category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}
