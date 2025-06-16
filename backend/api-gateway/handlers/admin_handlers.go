package handlers

import (
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

func BanUser(c *gin.Context) {
	log.Printf("BanUser: Handling ban user request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	userID := c.Param("userId")
	if userID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	// Parse the request body as a raw map to handle case-insensitive fields
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		log.Printf("BanUser Handler: Failed to parse request body: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}

	// Look for 'ban' or 'Ban' field in the raw data
	var ban bool
	var banFound bool

	// Try lowercase first
	if banValue, exists := rawData["ban"]; exists {
		if boolValue, ok := banValue.(bool); ok {
			ban = boolValue
			banFound = true
		}
	}

	// If not found, try uppercase
	if !banFound {
		if banValue, exists := rawData["Ban"]; exists {
			if boolValue, ok := banValue.(bool); ok {
				ban = boolValue
				banFound = true
			}
		}
	}

	if !banFound {
		log.Printf("BanUser Handler: 'ban' field not found in request: %v", rawData)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Missing required 'ban' field")
		return
	}

	adminID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin ID not found in token")
		return
	}

	// Extract reason if present
	var reason string
	if reasonValue, exists := rawData["reason"]; exists {
		if strValue, ok := reasonValue.(string); ok {
			reason = strValue
		}
	} else if reasonValue, exists := rawData["Reason"]; exists {
		if strValue, ok := reasonValue.(string); ok {
			reason = strValue
		}
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("BanUser Handler: Sending request to user service with ban=%v for user ID %s, reason: %s", ban, userID, reason)
	response, err := UserClient.BanUser(ctx, &userProto.BanUserRequest{
		UserId:  userID,
		Ban:     ban,
		Reason:  reason,
		AdminId: adminID.(string),
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("BanUser Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user ban status")
			}
		} else {
			log.Printf("BanUser Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user ban status")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": response.Message,
	})
}

func SendNewsletter(c *gin.Context) {
	log.Printf("SendNewsletter: Handling send newsletter request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	var req struct {
		Subject string `json:"subject" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("SendNewsletter Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	adminID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin not authenticated")
		return
	}
	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("SendNewsletter Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send newsletter")
			}
		} else {
			log.Printf("SendNewsletter Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send newsletter")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":          response.Message,
		"recipients_count": response.RecipientsCount,
	})
}

func GetCommunityRequests(c *gin.Context) {
	log.Printf("GetCommunityRequests: Handling get community requests endpoint")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

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

	if CommunityClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get unapproved communities directly from the community service
	isApproved := false
	communitiesResponse, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
		Query:      "",
		Limit:      int32(limit),
		Offset:     int32((page - 1) * limit),
		IsApproved: isApproved,
	})

	if err != nil {
		log.Printf("Error getting pending communities: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get pending communities")
		return
	}

	// Convert communities to community requests format
	communityRequests := make([]*userProto.CommunityRequest, 0, len(communitiesResponse.Communities))
	for _, community := range communitiesResponse.Communities {
		// Get creator info if available
		var requester *userProto.User
		if UserClient != nil && community.CreatorId != "" {
			userResp, userErr := UserClient.GetUser(ctx, &userProto.GetUserRequest{
				UserId: community.CreatorId,
			})
			if userErr == nil && userResp != nil && userResp.User != nil {
				requester = userResp.User
				log.Printf("Found creator %s for community %s", requester.Name, community.Name)
			} else {
				log.Printf("Could not find creator info for ID %s: %v", community.CreatorId, userErr)
			}
		}

		communityRequest := &userProto.CommunityRequest{
			Id:          community.Id,
			UserId:      community.CreatorId,
			Name:        community.Name,
			Description: community.Description,
			Status:      "pending",
			CreatedAt:   community.CreatedAt.AsTime().Format(time.RFC3339),
			UpdatedAt:   community.UpdatedAt.AsTime().Format(time.RFC3339),
			Requester:   requester,
			LogoUrl:     community.LogoUrl,
			BannerUrl:   community.BannerUrl,
		}
		communityRequests = append(communityRequests, communityRequest)
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"requests":    communityRequests,
		"total_count": communitiesResponse.TotalCount,
		"page":        int32(page),
		"limit":       int32(limit),
	})
}

func ProcessCommunityRequest(c *gin.Context) {
	log.Printf("ProcessCommunityRequest: Processing community request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	requestID := c.Param("requestId")
	if requestID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessCommunityRequest Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// The community ID is the same as the request ID
	communityID := requestID

	if CommunityClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Directly update the community approval status in the community service
	if req.Approve {
		_, approveErr := CommunityClient.ApproveCommunity(ctx, &communityProto.ApproveCommunityRequest{
			CommunityId: communityID,
		})

		if approveErr != nil {
			log.Printf("Error approving community %s: %v", communityID, approveErr)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to approve community")
			return
		}

		log.Printf("Successfully approved community %s", communityID)
	} else {
		// For rejection, we could implement a delete or mark as rejected in the future
		log.Printf("Community %s was rejected", communityID)
	}

	// Also update the community request status in the user service if it exists
	// This is for backward compatibility
	if UserClient != nil {
		_, err := UserClient.ProcessCommunityRequest(ctx, &userProto.ProcessCommunityRequestRequest{
			RequestId: requestID,
			Approve:   req.Approve,
		})

		if err != nil {
			log.Printf("Warning: Failed to update community request in user service: %v", err)
			// Continue even if this fails, as we've already updated the main community record
		} else {
			log.Printf("Successfully updated community request in user service")
		}
	}

	action := "rejected"
	if req.Approve {
		action = "approved"
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success": true,
		"message": "Community has been " + action,
	})
}

func GetPremiumRequests(c *gin.Context) {
	log.Printf("GetPremiumRequests: Handling premium requests endpoint")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

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
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get premium requests")
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

func ProcessPremiumRequest(c *gin.Context) {
	log.Printf("ProcessPremiumRequest: Processing premium request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	requestID := c.Param("requestId")
	if requestID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessPremiumRequest Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Premium request not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessPremiumRequest Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process premium request")
			}
		} else {
			log.Printf("ProcessPremiumRequest Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process premium request")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": response.Message,
	})
}

func GetReportRequests(c *gin.Context) {
	log.Printf("GetReportRequests: Handling report requests endpoint")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

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
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get report requests")
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

func ProcessReportRequest(c *gin.Context) {
	log.Printf("ProcessReportRequest: Processing report request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	requestID := c.Param("requestId")
	if requestID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Request ID is required")
		return
	}

	var req struct {
		Approve bool   `json:"approve" binding:"required"`
		Reason  string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ProcessReportRequest Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Report request not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessReportRequest Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process report request")
			}
		} else {
			log.Printf("ProcessReportRequest Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process report request")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": response.Message,
	})
}

func GetThreadCategories(c *gin.Context) {
	log.Printf("GetThreadCategories: Handling thread categories endpoint")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

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
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get thread categories")
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

func CreateThreadCategory(c *gin.Context) {
	log.Printf("CreateThreadCategory: Creating thread category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateThreadCategory Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("CreateThreadCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create thread category")
			}
		} else {
			log.Printf("CreateThreadCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create thread category")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

func UpdateThreadCategory(c *gin.Context) {
	log.Printf("UpdateThreadCategory: Updating thread category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateThreadCategory Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread category not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("UpdateThreadCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread category")
			}
		} else {
			log.Printf("UpdateThreadCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

func DeleteThreadCategory(c *gin.Context) {
	log.Printf("DeleteThreadCategory: Deleting thread category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread category not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("DeleteThreadCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete thread category")
			}
		} else {
			log.Printf("DeleteThreadCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete thread category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func GetCommunityCategories(c *gin.Context) {
	log.Printf("GetCommunityCategories: Handling community categories endpoint")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

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
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get community categories")
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

func CreateCommunityCategory(c *gin.Context) {
	log.Printf("CreateCommunityCategory: Creating community category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateCommunityCategory Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("CreateCommunityCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create community category")
			}
		} else {
			log.Printf("CreateCommunityCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create community category")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

func UpdateCommunityCategory(c *gin.Context) {
	log.Printf("UpdateCommunityCategory: Updating community category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateCommunityCategory Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community category not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("UpdateCommunityCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update community category")
			}
		} else {
			log.Printf("UpdateCommunityCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update community category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": response.Category,
	})
}

func DeleteCommunityCategory(c *gin.Context) {
	log.Printf("DeleteCommunityCategory: Deleting community category")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Category ID is required")
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community category not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("DeleteCommunityCategory Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete community category")
			}
		} else {
			log.Printf("DeleteCommunityCategory Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete community category")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func GetDashboardStatistics(c *gin.Context) {
	log.Printf("GetDashboardStatistics: Generating statistics data for admin dashboard")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"total_users":       int64(1250),
		"active_users":      int64(875),
		"total_communities": int64(45),
		"total_threads":     int64(3820),
		"pending_reports":   int64(12),
		"new_users_today":   int64(28),
		"new_posts_today":   int64(175),
	})
}

func AdminGetAllUsers(c *gin.Context) {
	log.Printf("AdminGetAllUsers: Handling get all users request")

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	page := 1
	limit := 10
	sortBy := "created_at"
	sortDesc := true

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

	if sortByStr := c.Query("sort_by"); sortByStr != "" {
		sortBy = sortByStr
	}

	if sortDescStr := c.Query("sort_desc"); sortDescStr != "" {
		if sd, err := strconv.ParseBool(sortDescStr); err == nil {
			sortDesc = sd
		}
	}
	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := UserClient.GetAllUsers(ctx, &userProto.GetAllUsersRequest{
		Page:           int32(page),
		Limit:          int32(limit),
		SortBy:         sortBy,
		SortDesc:       sortDesc,
		SearchQuery:    c.Query("search"),
		NewsletterOnly: false,
	})

	if err != nil {
		log.Printf("AdminGetAllUsers Handler: gRPC error: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"users":       response.Users,
		"total_count": response.TotalCount,
		"page":        response.Page,
	})
}

// SyncPendingCommunities fetches communities from the community service that are not approved
// and returns statistics about them
func SyncPendingCommunities(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all pending communities from community service
	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	// Set false to filter by not approved communities
	isApproved := false
	searchResp, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
		IsApproved: isApproved,
		Limit:      100, // Using a reasonable limit
		Offset:     0,
	})

	if err != nil {
		log.Printf("Error getting pending communities: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get pending communities")
		return
	}

	pendingCommunities := searchResp.Communities

	// Track results
	var syncResults struct {
		TotalPendingCommunities int      `json:"total_pending_communities"`
		PendingCommunityIds     []string `json:"pending_community_ids,omitempty"`
		CreatorIds              []string `json:"creator_ids,omitempty"`
	}

	syncResults.TotalPendingCommunities = len(pendingCommunities)

	// Collect IDs for debugging
	for _, community := range pendingCommunities {
		syncResults.PendingCommunityIds = append(syncResults.PendingCommunityIds, community.Id)
		if community.CreatorId != "" {
			syncResults.CreatorIds = append(syncResults.CreatorIds, community.CreatorId)
		}
	}

	utils.SendSuccessResponse(c, 200, syncResults)
}
