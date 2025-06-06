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

	// Add CORS headers
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

	var req struct {
		Ban    bool   `json:"ban" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("BanUser Handler: Invalid request payload: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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

	// Add CORS headers
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
	log.Printf("GetCommunityRequests: Handling community requests endpoint")

	// Add CORS headers
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

	response, err := UserClient.GetCommunityRequests(ctx, &userProto.GetCommunityRequestsRequest{
		Page:   int32(page),
		Limit:  int32(limit),
		Status: status,
	})

	if err != nil {
		log.Printf("GetCommunityRequests Handler: gRPC error: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get community requests")
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

func ProcessCommunityRequest(c *gin.Context) {
	log.Printf("ProcessCommunityRequest: Processing community request")

	// Add CORS headers
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

	if UserClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
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
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Community request not found")
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
			default:
				log.Printf("ProcessCommunityRequest Handler: gRPC error: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process community request")
			}
		} else {
			log.Printf("ProcessCommunityRequest Handler: Unknown error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process community request")
		}
		return
	}

	// If the community request was approved, also update the community's approved status
	if req.Approve && CommunityClient != nil && response.Success {
		// Get the community ID directly from the community service
		// Since naming is consistent, the community ID should be the same as the Name in the request
		communityRequestsResponse, err := UserClient.GetCommunityRequests(ctx, &userProto.GetCommunityRequestsRequest{
			Page:   1,
			Limit:  100,
			Status: "approved", // Look for the request we just approved
		})

		if err == nil && communityRequestsResponse.Requests != nil && len(communityRequestsResponse.Requests) > 0 {
			// Find the request we just processed
			var communityName string
			for _, request := range communityRequestsResponse.Requests {
				if request.Id == requestID {
					communityName = request.Name
					break
				}
			}

			if communityName != "" {
				// Search for the community by name in the community service
				searchResponse, searchErr := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
					Query:  communityName,
					Limit:  1,
					Offset: 0,
				})

				if searchErr == nil && searchResponse.Communities != nil && len(searchResponse.Communities) > 0 {
					communityID := searchResponse.Communities[0].Id

					// Call the community service to approve the community
					_, approveErr := CommunityClient.ApproveCommunity(ctx, &communityProto.ApproveCommunityRequest{
						CommunityId: communityID,
					})

					if approveErr != nil {
						// Log error but don't fail the request, as the request itself was processed successfully
						log.Printf("Warning: Failed to approve community in community service: %v", approveErr)
					} else {
						log.Printf("Successfully approved community %s in community service", communityID)
					}
				} else {
					log.Printf("Warning: Could not find community with name '%s' in community service: %v", communityName, searchErr)
				}
			} else {
				log.Printf("Warning: Could not find community request details for ID %s", requestID)
			}
		} else {
			log.Printf("Warning: Failed to retrieve community requests: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func GetPremiumRequests(c *gin.Context) {
	log.Printf("GetPremiumRequests: Handling premium requests endpoint")

	// Add CORS headers
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

	// Add CORS headers
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

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func GetReportRequests(c *gin.Context) {
	log.Printf("GetReportRequests: Handling report requests endpoint")

	// Add permissive CORS headers
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

	// Add permissive CORS headers
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

	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
	})
}

func GetThreadCategories(c *gin.Context) {
	log.Printf("GetThreadCategories: Handling thread categories endpoint")

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
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

	// Add CORS headers
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	// For now, we use placeholder values until the API is fully implemented
	// In a production environment, these would come from actual database queries
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
