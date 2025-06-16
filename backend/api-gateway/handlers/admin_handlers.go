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

	for i, req := range response.Requests {
		if req.Requester == nil && req.UserId != "" {
			log.Printf("Request %s missing requester info, fetching user data for ID: %s", req.Id, req.UserId)
			userResp, userErr := UserClient.GetUser(ctx, &userProto.GetUserRequest{
				UserId: req.UserId,
			})
			if userErr == nil && userResp != nil && userResp.User != nil {
				response.Requests[i].Requester = userResp.User
				log.Printf("Added missing requester %s for community request %s", userResp.User.Name, req.Name)
			} else {
				log.Printf("Warning: Could not find user %s for adding to community request: %v", req.UserId, userErr)
			}
		} else if req.Requester != nil {
			log.Printf("Request %s already has requester info: %s", req.Id, req.Requester.Name)
		} else {
			log.Printf("Request %s has no user_id to fetch requester for", req.Id)
		}
	}

	if CommunityClient != nil {
		communitiesResponse, commErr := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
			Query:      "",
			Limit:      int32(limit),
			Offset:     int32((page - 1) * limit),
			IsApproved: false,
		})

		if commErr == nil && communitiesResponse != nil && len(communitiesResponse.Communities) > 0 {
			log.Printf("Found %d communities from community service", len(communitiesResponse.Communities))

			// Create a map to track existing requests by name for faster lookup
			existingRequests := make(map[string]bool)
			for _, req := range response.Requests {
				existingRequests[req.Name] = true
			}

			for _, community := range communitiesResponse.Communities {
				// Check if we already have a request with this name using our map
				if !existingRequests[community.Name] && community.CreatorId != "" {
					log.Printf("Adding community %s with creator ID: %s", community.Name, community.CreatorId)

					var requester *userProto.User
					if UserClient != nil {
						log.Printf("Fetching user info for creator ID: %s", community.CreatorId)

						userResp, userErr := UserClient.GetUser(ctx, &userProto.GetUserRequest{
							UserId: community.CreatorId,
						})

						if userErr != nil {
							log.Printf("ERROR fetching creator info: %v", userErr)
						} else if userResp == nil {
							log.Printf("ERROR: GetUser returned nil response for ID: %s", community.CreatorId)
						} else if userResp.User == nil {
							log.Printf("ERROR: GetUser returned nil user for ID: %s", community.CreatorId)
						} else {
							requester = userResp.User
							log.Printf("SUCCESS: Found creator %s (%s) for community %s",
								requester.Name, requester.Username, community.Name)
						}
					} else {
						log.Printf("ERROR: UserClient is nil when trying to fetch creator info")
					}

					// Add the community as a pending request
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

					if requester != nil {
						log.Printf("Added community request with requester: %s (%s)", requester.Name, requester.Username)
					} else {
						log.Printf("WARNING: Added community request WITHOUT requester for ID: %s", community.CreatorId)
					}

					response.Requests = append(response.Requests, communityRequest)
					response.TotalCount++
				}
			}
		} else if commErr != nil {
			log.Printf("ERROR: Failed to get communities: %v", commErr)
		}
	}

	for i, req := range response.Requests {
		if req.Requester == nil && req.UserId != "" {
			log.Printf("Final check: Request %s for community %s is missing requester info, fetching user data for ID: %s",
				req.Id, req.Name, req.UserId)

			if UserClient == nil {
				log.Printf("ERROR: UserClient is nil during final requester check")
				continue
			}

			retryCount := 2
			var userResp *userProto.GetUserResponse
			var userErr error

			for attempt := 0; attempt <= retryCount; attempt++ {
				if attempt > 0 {
					log.Printf("Retry attempt %d for fetching user %s", attempt, req.UserId)
				}

				userResp, userErr = UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: req.UserId,
				})

				if userErr == nil && userResp != nil && userResp.User != nil {
					response.Requests[i].Requester = userResp.User
					log.Printf("SUCCESS: Added requester %s (%s) for community %s",
						userResp.User.Name, userResp.User.Username, req.Name)
					break
				}

				if attempt < retryCount {
					time.Sleep(100 * time.Millisecond)
				}
			}

			if userErr != nil || userResp == nil || userResp.User == nil {

				if userErr != nil {
					log.Printf("ERROR: Failed to fetch user %s after %d attempts: %v", req.UserId, retryCount+1, userErr)
				} else if userResp == nil {
					log.Printf("ERROR: Empty response when fetching user %s after %d attempts", req.UserId, retryCount+1)
				} else {
					log.Printf("ERROR: User object is nil in response for ID %s after %d attempts", req.UserId, retryCount+1)
				}
			}
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"requests":    response.Requests,
		"total_count": response.TotalCount,
		"page":        response.Page,
		"limit":       response.Limit,
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

	// If the request was approved, also update the community in the community service
	if req.Approve && CommunityClient != nil && response.Success {
		// The request ID should be the same as the community ID
		communityID := requestID

		// Call the ApproveCommunity endpoint to update the community's approval status
		_, approveErr := CommunityClient.ApproveCommunity(ctx, &communityProto.ApproveCommunityRequest{
			CommunityId: communityID,
		})

		if approveErr != nil {
			log.Printf("Warning: Failed to approve community in community service: %v", approveErr)
		} else {
			log.Printf("Successfully approved community %s in community service", communityID)
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success": response.Success,
		"message": response.Message,
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
// and ensures they exist in the user service's community_requests table
func SyncPendingCommunities(c *gin.Context) {
	// Verify admin permissions
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check if user is admin
	if UserClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "User service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userResp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID.(string),
	})
	if err != nil || !userResp.User.IsAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Admin privileges required")
		return
	}

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

	// Get existing community requests from user service
	communityReqsResp, err := UserClient.GetCommunityRequests(ctx, &userProto.GetCommunityRequestsRequest{
		Limit: 1000, // High limit to get as many as possible
		Page:  1,
	})

	if err != nil {
		log.Printf("Error getting existing community requests: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get existing community requests")
		return
	}

	// Create a map of existing community request IDs for quick lookup
	existingRequestsMap := make(map[string]bool)
	for _, req := range communityReqsResp.Requests {
		existingRequestsMap[req.Id] = true
	}

	// Track results
	var syncResults struct {
		TotalPendingCommunities int      `json:"total_pending_communities"`
		AlreadySynced           int      `json:"already_synced"`
		NewlySynced             int      `json:"newly_synced"`
		Failed                  int      `json:"failed"`
		FailedCommunityIds      []string `json:"failed_community_ids,omitempty"`
	}

	syncResults.TotalPendingCommunities = len(pendingCommunities)

	// For each pending community not already in the user service, create a request
	for _, community := range pendingCommunities {
		// Skip if already exists in user service
		if existingRequestsMap[community.Id] {
			syncResults.AlreadySynced++
			continue
		}

		// Get detailed community info if needed
		communityDetail, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: community.Id,
		})

		if err != nil {
			log.Printf("Error getting community details for ID %s: %v", community.Id, err)
			syncResults.Failed++
			syncResults.FailedCommunityIds = append(syncResults.FailedCommunityIds, community.Id)
			continue
		}

		// Create community request in user service
		_, err = UserClient.CreateCommunityRequest(ctx, &userProto.CreateCommunityRequestRequest{
			CommunityId: community.Id,
			UserId:      communityDetail.Community.CreatorId,
			Name:        community.Name,
			Description: community.Description,
		})

		if err != nil {
			log.Printf("Error creating community request for community %s: %v", community.Id, err)
			syncResults.Failed++
			syncResults.FailedCommunityIds = append(syncResults.FailedCommunityIds, community.Id)
		} else {
			log.Printf("Successfully synced community request for community ID: %s", community.Id)
			syncResults.NewlySynced++
		}
	}

	utils.SendSuccessResponse(c, 200, syncResults)
}
