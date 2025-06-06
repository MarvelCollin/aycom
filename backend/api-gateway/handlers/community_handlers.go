package handlers

import (
	"aycom/backend/api-gateway/utils"
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"
	"context"
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check content type to determine if it's JSON or multipart form
	contentType := c.GetHeader("Content-Type")
	var name, description, logoURL, bannerURL, rules string
	var categories []string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart form data
		log.Printf("Handling multipart form data")
		name = c.PostForm("name")
		description = c.PostForm("description")
		rules = c.PostForm("rules")

		// Get the categories
		categoriesJSON := c.PostForm("categories")
		if categoriesJSON != "" {
			if err := json.Unmarshal([]byte(categoriesJSON), &categories); err != nil {
				log.Printf("CreateCommunity: Invalid categories format: %v", err)
				utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid categories format")
				return
			}
		}

		// Handle file uploads using Supabase
		// Get logo file
		logoFile, err := c.FormFile("icon")
		if err != nil {
			log.Printf("CreateCommunity: No logo file: %v", err)
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Logo file is required")
			return
		}

		// Get banner file
		bannerFile, err := c.FormFile("banner")
		if err != nil {
			log.Printf("CreateCommunity: No banner file: %v", err)
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Banner file is required")
			return
		}

		// Open logo file
		logoFileOpen, err := logoFile.Open()
		if err != nil {
			log.Printf("CreateCommunity: Failed to open logo file: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to process logo file")
			return
		}
		defer logoFileOpen.Close()

		// Open banner file
		bannerFileOpen, err := bannerFile.Open()
		if err != nil {
			log.Printf("CreateCommunity: Failed to open banner file: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to process banner file")
			return
		}
		defer bannerFileOpen.Close()

		// Upload logo to Supabase
		logoURL, err = utils.UploadFile(logoFileOpen, logoFile.Filename, "media", "communities/logos")
		if err != nil {
			log.Printf("CreateCommunity: Failed to upload logo to Supabase: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to upload logo")
			return
		}

		// Upload banner to Supabase
		bannerURL, err = utils.UploadFile(bannerFileOpen, bannerFile.Filename, "media", "communities/banners")
		if err != nil {
			log.Printf("CreateCommunity: Failed to upload banner to Supabase: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to upload banner")
			return
		}

		// Validate required fields
		if name == "" || description == "" || rules == "" || len(categories) == 0 {
			log.Printf("CreateCommunity: Missing required fields")
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Missing required fields")
			return
		}
	} else {
		// Handle JSON
		var req struct {
			Name        string   `json:"name" binding:"required"`
			Description string   `json:"description" binding:"required"`
			LogoURL     string   `json:"logo_url" binding:"required"`
			BannerURL   string   `json:"banner_url" binding:"required"`
			Rules       string   `json:"rules" binding:"required"`
			Categories  []string `json:"categories" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("CreateCommunity: Invalid request body: %v", err)
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
			return
		}

		name = req.Name
		description = req.Description
		logoURL = req.LogoURL
		bannerURL = req.BannerURL
		rules = req.Rules
		categories = req.Categories
	}

	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create categories array for proto
	protoCategories := make([]*communityProto.Category, len(categories))
	for i, categoryName := range categories {
		protoCategories[i] = &communityProto.Category{
			Name: categoryName,
		}
	}

	// Create the community
	community := &communityProto.Community{
		Name:        name,
		Description: description,
		LogoUrl:     logoURL,
		BannerUrl:   bannerURL,
		CreatorId:   userID.(string),
		IsApproved:  false, // New communities are not auto-approved
		Categories:  protoCategories,
	}

	resp, err := CommunityClient.CreateCommunity(ctx, &communityProto.CreateCommunityRequest{
		Community: community,
	})

	if err != nil {
		log.Printf("Error calling CreateCommunity: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to create community: "+err.Error())
		return
	}

	// Create community rules
	if resp != nil && resp.Community != nil {
		_, err = CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
			CommunityId: resp.Community.Id,
			RuleText:    rules,
		})

		if err != nil {
			log.Printf("Error adding rules to community: %v", err)
			// Don't fail the entire request if just rules failed
		}
	}

	// Add creator as admin member
	if resp != nil && resp.Community != nil {
		_, err = CommunityClient.AddMember(ctx, &communityProto.AddMemberRequest{
			CommunityId: resp.Community.Id,
			UserId:      userID.(string),
			Role:        "admin",
		})

		if err != nil {
			log.Printf("Error adding creator as member: %v", err)
			// Don't fail the entire request if just adding member failed
		}

		// Also create a community request in the user service
		if UserClient != nil {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			_, err = UserClient.CreateCommunityRequest(userCtx, &userProto.CreateCommunityRequestRequest{
				CommunityId: resp.Community.Id,
				UserId:      userID.(string),
				Name:        name,
				Description: description,
			})

			if err != nil {
				log.Printf("Error creating community request in user service: %v", err)
				// Don't fail the entire request if just the request creation failed
			} else {
				log.Printf("Successfully created community request in user service for community ID: %s", resp.Community.Id)
			}
		} else {
			log.Printf("WARNING: UserClient is nil! Could not create community request in user service")
		}
	}

	// Format the response
	formattedCategories := make([]string, 0)
	if resp.Community.Categories != nil {
		for _, cat := range resp.Community.Categories {
			formattedCategories = append(formattedCategories, cat.Name)
		}
	}

	createdAt := time.Now()
	if resp.Community.CreatedAt != nil {
		createdAt = resp.Community.CreatedAt.AsTime()
	}

	communityData := gin.H{
		"id":           resp.Community.Id,
		"name":         resp.Community.Name,
		"description":  resp.Community.Description,
		"logo_url":     resp.Community.LogoUrl,
		"banner_url":   resp.Community.BannerUrl,
		"creator_id":   resp.Community.CreatorId,
		"is_approved":  resp.Community.IsApproved,
		"categories":   formattedCategories,
		"created_at":   createdAt,
		"member_count": 1, // Initially only the creator is a member
	}

	utils.SendSuccessResponse(c, 201, communityData)
}

func UpdateCommunity(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description" binding:"required"`
		LogoURL     string   `json:"logo_url" binding:"required"`
		BannerURL   string   `json:"banner_url" binding:"required"`
		Rules       string   `json:"rules"`
		Categories  []string `json:"categories"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateCommunity: Invalid request body: %v", err)
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current community details
	getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error getting community for update: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
		return
	}

	if getCommunityResp.Community.CreatorId != userID.(string) {
		// Check if user is admin
		isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
			CommunityId: communityID,
			UserId:      userID.(string),
		})

		if err != nil || !isMemberResp.IsMember {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community creator or admin can update the community")
			return
		}

		// Check if user is admin by getting their role
		membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
			CommunityId: communityID,
		})

		isAdmin := false
		if err == nil && membersResp.Members != nil {
			for _, member := range membersResp.Members {
				if member.UserId == userID.(string) && member.Role == "admin" {
					isAdmin = true
					break
				}
			}
		}

		if !isAdmin {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community creator or admin can update the community")
			return
		}
	}

	// Create categories array for proto
	categories := make([]*communityProto.Category, len(req.Categories))
	for i, categoryName := range req.Categories {
		categories[i] = &communityProto.Category{
			Name: categoryName,
		}
	}

	// Update the community
	community := &communityProto.Community{
		Id:          communityID,
		Name:        req.Name,
		Description: req.Description,
		LogoUrl:     req.LogoURL,
		BannerUrl:   req.BannerURL,
		Categories:  categories,
	}

	resp, err := CommunityClient.UpdateCommunity(ctx, &communityProto.UpdateCommunityRequest{
		Community: community,
	})

	if err != nil {
		log.Printf("Error calling UpdateCommunity: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to update community: "+err.Error())
		return
	}

	// Update community rules if provided
	if req.Rules != "" {
		// First get existing rules
		rulesResp, err := CommunityClient.ListRules(ctx, &communityProto.ListRulesRequest{
			CommunityId: communityID,
		})

		// Delete existing rules
		if err == nil && rulesResp.Rules != nil {
			for _, rule := range rulesResp.Rules {
				_, _ = CommunityClient.RemoveRule(ctx, &communityProto.RemoveRuleRequest{
					RuleId: rule.Id,
				})
			}
		}

		// Add new rule
		_, err = CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
			CommunityId: communityID,
			RuleText:    req.Rules,
		})

		if err != nil {
			log.Printf("Error updating rules: %v", err)
			// Don't fail the entire request if just rules failed
		}
	}

	// Format the response
	formattedCategories := make([]string, 0)
	if resp.Community.Categories != nil {
		for _, cat := range resp.Community.Categories {
			formattedCategories = append(formattedCategories, cat.Name)
		}
	}

	createdAt := time.Now()
	if resp.Community.CreatedAt != nil {
		createdAt = resp.Community.CreatedAt.AsTime()
	}

	communityData := gin.H{
		"id":           resp.Community.Id,
		"name":         resp.Community.Name,
		"description":  resp.Community.Description,
		"logo_url":     resp.Community.LogoUrl,
		"banner_url":   resp.Community.BannerUrl,
		"creator_id":   resp.Community.CreatorId,
		"is_approved":  resp.Community.IsApproved,
		"categories":   formattedCategories,
		"created_at":   createdAt,
		"member_count": 0, // Need to fetch member count separately
	}

	utils.SendSuccessResponse(c, 200, communityData)
}

func ApproveCommunity(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	// Check if user is an admin
	// Here you would typically check if the user is a system admin
	// This is a simplified check - in a real app, you'd check against admin roles
	var isAdmin bool
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only administrators can approve communities")
		return
	}

	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ApproveCommunity(ctx, &communityProto.ApproveCommunityRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling ApproveCommunity: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to approve community: "+err.Error())
		return
	}

	// Format the response
	formattedCategories := make([]string, 0)
	if resp.Community.Categories != nil {
		for _, cat := range resp.Community.Categories {
			formattedCategories = append(formattedCategories, cat.Name)
		}
	}

	createdAt := time.Now()
	if resp.Community.CreatedAt != nil {
		createdAt = resp.Community.CreatedAt.AsTime()
	}

	communityData := gin.H{
		"id":           resp.Community.Id,
		"name":         resp.Community.Name,
		"description":  resp.Community.Description,
		"logo_url":     resp.Community.LogoUrl,
		"banner_url":   resp.Community.BannerUrl,
		"creator_id":   resp.Community.CreatorId,
		"is_approved":  resp.Community.IsApproved,
		"categories":   formattedCategories,
		"created_at":   createdAt,
		"member_count": 0,
	}

	utils.SendSuccessResponse(c, 200, communityData)
}

func DeleteCommunity(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get community to check if user is the creator
	getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error getting community for deletion: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
		return
	}

	// Check if user is the creator or an admin
	if getCommunityResp.Community.CreatorId != userID.(string) {
		// Check if user is admin
		isAdmin := false
		adminIDStr, adminExists := c.Get("isAdmin")
		if adminExists && adminIDStr.(bool) {
			isAdmin = true
		}

		// If not admin, check if they're a community admin
		if !isAdmin {
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}

		if !isAdmin {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community creator or admins can delete the community")
			return
		}
	}

	_, err = CommunityClient.DeleteCommunity(ctx, &communityProto.DeleteCommunityRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling DeleteCommunity: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to delete community: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Community deleted successfully",
	})
}

func GetCommunityByID(c *gin.Context) {
	communityID := c.Param("id")
	log.Printf("GetCommunityByID called with ID: %s", communityID)

	if communityID == "" {
		log.Printf("Error: Empty community ID provided")
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling GetCommunityByID: %v", err)

		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			utils.SendErrorResponse(c, 404, "NOT_FOUND", "Community not found")
		} else {
			utils.SendErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to get community: "+err.Error())
		}
		return
	}

	if resp == nil || resp.Community == nil {
		log.Printf("GetCommunityByID returned nil response or nil community")
		utils.SendErrorResponse(c, 404, "NOT_FOUND", "Community not found")
		return
	}

	community := resp.Community

	formattedCategories := make([]string, 0)

	if community.Categories != nil {
		for _, cat := range community.Categories {
			formattedCategories = append(formattedCategories, cat.Name)
		}
	}

	createdAt := time.Now()
	if community.CreatedAt != nil {
		createdAt = community.CreatedAt.AsTime()
	}

	// Default member count to 0
	memberCount := int64(0)

	// We can add actual member count logic here when implemented in the proto

	communityData := gin.H{
		"id":           community.Id,
		"name":         community.Name,
		"description":  community.Description,
		"logo_url":     community.LogoUrl,
		"banner_url":   community.BannerUrl,
		"creator_id":   community.CreatorId,
		"is_approved":  community.IsApproved,
		"categories":   formattedCategories,
		"created_at":   createdAt,
		"member_count": memberCount,
	}

	utils.SendSuccessResponse(c, 200, communityData)
}

func ListCommunities(c *gin.Context) {

	limitOptions := []int{25, 30, 35}
	limit := limitOptions[0]

	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil {

			validLimit := false
			for _, option := range limitOptions {
				if parsedLimit == option {
					validLimit = true
					limit = parsedLimit
					break
				}
			}

			if !validLimit {
				limit = limitOptions[0]
			}
		}
	}

	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	offset := (page - 1) * limit

	_ = c.DefaultQuery("filter", "all")
	_ = c.Query("q")
	_ = c.QueryArray("category")

	var communities []*communityProto.Community
	var totalCount int32 = 0

	if CommunityClient == nil {
		log.Printf("CommunityClient is nil")
		utils.SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
		Offset: int32(offset),
		Limit:  int32(limit),
	})

	if err != nil {
		log.Printf("Error calling ListCommunities: %v", err)
		utils.SendErrorResponse(c, 500, "server_error", "Failed to list communities: "+err.Error())
		return
	}

	communities = resp.GetCommunities()
	totalCount = resp.GetTotalCount()

	formattedCommunities := make([]gin.H, 0, len(communities))
	for _, comm := range communities {
		// Check for categories in this community
		formattedCategories := []string{}
		if comm.Categories != nil {
			for _, cat := range comm.Categories {
				formattedCategories = append(formattedCategories, cat.Name)
			}
		}

		createdAt := time.Now()
		if comm.CreatedAt != nil {
			createdAt = comm.CreatedAt.AsTime()
		}

		formattedCommunities = append(formattedCommunities, gin.H{
			"id":          comm.Id,
			"name":        comm.Name,
			"description": comm.Description,
			"logo_url":    comm.LogoUrl,
			"banner_url":  comm.BannerUrl,
			"creator_id":  comm.CreatorId,
			"is_approved": comm.IsApproved,
			"created_at":  createdAt,
			"categories":  formattedCategories,
		})
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": formattedCommunities,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
	})
}

func ListCategories(c *gin.Context) {
	if CommunityClient == nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListCategories(ctx, &communityProto.ListCategoriesRequest{})
	if err != nil {
		log.Printf("Error calling ListCategories: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list categories: "+err.Error())
		return
	}

	formattedCategories := make([]gin.H, 0, len(resp.Categories))
	for _, category := range resp.Categories {
		formattedCategories = append(formattedCategories, gin.H{
			"id":   category.Id,
			"name": category.Name,
		})
	}

	// If no categories returned from service, provide default ones
	if len(formattedCategories) == 0 {
		formattedCategories = []gin.H{
			{"id": "1", "name": "Technology"},
			{"id": "2", "name": "Gaming"},
			{"id": "3", "name": "Education"},
			{"id": "4", "name": "Entertainment"},
			{"id": "5", "name": "Sports"},
			{"id": "6", "name": "Business"},
			{"id": "7", "name": "Art"},
			{"id": "8", "name": "Science"},
			{"id": "9", "name": "Health"},
			{"id": "10", "name": "Music"},
		}
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"categories": formattedCategories,
	})
}

func AddMember(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if req.Role == "" {
		req.Role = "member" // Default role
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can add members")
		return
	}

	// Check if user already exists in community
	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      req.UserID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check membership: "+err.Error())
		return
	}

	if isMemberResp.IsMember {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "User is already a member of this community")
		return
	}

	// Add the member
	resp, err := CommunityClient.AddMember(ctx, &communityProto.AddMemberRequest{
		CommunityId: communityID,
		UserId:      req.UserID,
		Role:        req.Role,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to add member: "+err.Error())
		return
	}

	joinedAt := time.Now()
	if resp.Member.JoinedAt != nil {
		joinedAt = resp.Member.JoinedAt.AsTime()
	}

	utils.SendSuccessResponse(c, 201, gin.H{
		"id":                  resp.Member.UserId,
		"user_id":             resp.Member.UserId,
		"username":            "user_" + resp.Member.UserId, // This would typically come from a user service
		"name":                "User " + resp.Member.UserId, // This would typically come from a user service
		"role":                resp.Member.Role,
		"joined_at":           joinedAt,
		"profile_picture_url": "", // This would typically come from a user service
	})
}

func RemoveMember(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	memberUserID := c.Param("userId")
	if memberUserID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Member User ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user to be removed is the authenticated user (self-removal)
	selfRemoval := memberUserID == userID.(string)

	// If not self-removal, check if user has admin permissions
	if !selfRemoval {
		isAdmin := false
		adminIDStr, adminExists := c.Get("isAdmin")
		if adminExists && adminIDStr.(bool) {
			isAdmin = true
		}

		if !isAdmin {
			// Get community to check if user is the creator
			getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
				CommunityId: communityID,
			})

			if err != nil {
				utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
				return
			}

			if getCommunityResp.Community.CreatorId == userID.(string) {
				isAdmin = true
			} else {
				// Check if user is a community admin
				membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
					CommunityId: communityID,
				})

				if err == nil && membersResp.Members != nil {
					for _, member := range membersResp.Members {
						if member.UserId == userID.(string) && member.Role == "admin" {
							isAdmin = true
							break
						}
					}
				}
			}
		}

		if !isAdmin {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can remove other members")
			return
		}
	}

	// Check if user is a member of the community
	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      memberUserID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check membership: "+err.Error())
		return
	}

	if !isMemberResp.IsMember {
		utils.SendErrorResponse(c, 404, "NOT_FOUND", "User is not a member of this community")
		return
	}

	// Remove the member
	_, err = CommunityClient.RemoveMember(ctx, &communityProto.RemoveMemberRequest{
		CommunityId: communityID,
		UserId:      memberUserID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to remove member: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Member removed successfully",
	})
}

func ListMembers(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "bad_request", "Community ID is required")
		return
	}

	limit := 20
	offset := 0

	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if pageParam := c.Query("page"); pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err == nil && parsedPage > 0 {
			offset = (parsedPage - 1) * limit
		}
	}

	if CommunityClient == nil {
		log.Printf("CommunityClient is nil")
		utils.SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling ListMembers: %v", err)

		if status.Code(err) == codes.NotFound {
			utils.SendErrorResponse(c, 404, "not_found", "Community not found")
			return
		}

		utils.SendErrorResponse(c, 500, "server_error", "Failed to list members: "+err.Error())
		return
	}

	formattedMembers := make([]gin.H, 0)
	if resp != nil && resp.Members != nil {
		for _, member := range resp.Members {
			joinedAt := time.Now()
			if member.JoinedAt != nil {
				joinedAt = member.JoinedAt.AsTime()
			}

			formattedMembers = append(formattedMembers, gin.H{
				"id":                  member.UserId,
				"user_id":             member.UserId,
				"username":            "user_" + member.UserId,
				"name":                "User " + member.UserId,
				"role":                member.Role,
				"joined_at":           joinedAt,
				"profile_picture_url": "",
			})
		}
	}

	totalCount := int32(len(formattedMembers))

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	currentPage := offset/limit + 1

	utils.SendSuccessResponse(c, 200, gin.H{
		"members": formattedMembers,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": currentPage,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
	})
}

func UpdateMemberRole(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	memberUserID := c.Param("userId")
	if memberUserID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Member User ID is required")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can update member roles")
		return
	}

	// Check if user is a member of the community
	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      memberUserID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check membership: "+err.Error())
		return
	}

	if !isMemberResp.IsMember {
		utils.SendErrorResponse(c, 404, "NOT_FOUND", "User is not a member of this community")
		return
	}

	// Update the member role
	resp, err := CommunityClient.UpdateMemberRole(ctx, &communityProto.UpdateMemberRoleRequest{
		CommunityId: communityID,
		UserId:      memberUserID,
		Role:        req.Role,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to update member role: "+err.Error())
		return
	}

	joinedAt := time.Now()
	if resp.Member.JoinedAt != nil {
		joinedAt = resp.Member.JoinedAt.AsTime()
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"id":                  resp.Member.UserId,
		"user_id":             resp.Member.UserId,
		"username":            "user_" + resp.Member.UserId, // This would typically come from a user service
		"name":                "User " + resp.Member.UserId, // This would typically come from a user service
		"role":                resp.Member.Role,
		"joined_at":           joinedAt,
		"profile_picture_url": "", // This would typically come from a user service
	})
}

func AddRule(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	var req struct {
		RuleText string `json:"rule_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can add rules")
		return
	}

	// Add the rule
	resp, err := CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
		CommunityId: communityID,
		RuleText:    req.RuleText,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to add rule: "+err.Error())
		return
	}

	// Rules typically have a display order, but we'll assign this on the client side
	// based on the order received from the server
	utils.SendSuccessResponse(c, 201, gin.H{
		"id":           resp.Rule.Id,
		"community_id": resp.Rule.CommunityId,
		"title":        "Community Rule",
		"description":  resp.Rule.RuleText,
		"order":        1, // Default order
	})
}

func RemoveRule(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	ruleID := c.Param("ruleId")
	if ruleID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Rule ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can remove rules")
		return
	}

	// Remove the rule
	_, err := CommunityClient.RemoveRule(ctx, &communityProto.RemoveRuleRequest{
		RuleId: ruleID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to remove rule: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Rule removed successfully",
	})
}

func ListRules(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("CommunityClient is nil")
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListRules(ctx, &communityProto.ListRulesRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling ListRules: %v", err)

		if status.Code(err) == codes.NotFound {
			utils.SendErrorResponse(c, 404, "NOT_FOUND", "Community not found")
			return
		}

		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list rules: "+err.Error())
		return
	}

	formattedRules := make([]gin.H, 0)
	if resp != nil && resp.Rules != nil {
		for i, rule := range resp.Rules {
			formattedRules = append(formattedRules, gin.H{
				"id":           rule.Id,
				"community_id": rule.CommunityId,
				"title":        "Rule " + strconv.Itoa(i+1),
				"description":  rule.RuleText,
				"order":        i + 1,
			})
		}
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"rules": formattedRules,
	})
}

func RequestToJoin(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.RequestToJoin(ctx, &communityProto.RequestToJoinRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.AlreadyExists {
			utils.SendErrorResponse(c, 400, "ALREADY_REQUESTED", "You have already requested to join this community")
			return
		}
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to request to join: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request sent successfully",
		"join_request": gin.H{
			"id":           resp.JoinRequest.Id,
			"community_id": resp.JoinRequest.CommunityId,
			"user_id":      resp.JoinRequest.UserId,
			"status":       resp.JoinRequest.Status,
		},
	})
}

func ApproveJoinRequest(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	requestID := c.Param("requestId")
	if requestID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Request ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can approve join requests")
		return
	}

	// Approve the join request
	resp, err := CommunityClient.ApproveJoinRequest(ctx, &communityProto.ApproveJoinRequestRequest{
		JoinRequestId: requestID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to approve join request: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request approved successfully",
		"join_request": gin.H{
			"id":           resp.JoinRequest.Id,
			"community_id": resp.JoinRequest.CommunityId,
			"user_id":      resp.JoinRequest.UserId,
			"status":       resp.JoinRequest.Status,
		},
	})
}
func RejectJoinRequest(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	requestID := c.Param("requestId")
	if requestID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Request ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can reject join requests")
		return
	}

	// Reject the join request
	resp, err := CommunityClient.RejectJoinRequest(ctx, &communityProto.RejectJoinRequestRequest{
		JoinRequestId: requestID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to reject join request: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request rejected successfully",
		"join_request": gin.H{
			"id":           resp.JoinRequest.Id,
			"community_id": resp.JoinRequest.CommunityId,
			"user_id":      resp.JoinRequest.UserId,
			"status":       resp.JoinRequest.Status,
		},
	})
}
func ListJoinRequests(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the user is admin of the community or system admin
	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {
		// Get community to check if user is the creator
		getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
			CommunityId: communityID,
		})

		if err != nil {
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
			return
		}

		if getCommunityResp.Community.CreatorId == userID.(string) {
			isAdmin = true
		} else {
			// Check if user is a community admin
			membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
				CommunityId: communityID,
			})

			if err == nil && membersResp.Members != nil {
				for _, member := range membersResp.Members {
					if member.UserId == userID.(string) && member.Role == "admin" {
						isAdmin = true
						break
					}
				}
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can view join requests")
		return
	}

	// List join requests
	resp, err := CommunityClient.ListJoinRequests(ctx, &communityProto.ListJoinRequestsRequest{
		CommunityId: communityID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list join requests: "+err.Error())
		return
	}

	formattedRequests := make([]gin.H, 0, len(resp.JoinRequests))
	for _, req := range resp.JoinRequests {
		formattedRequests = append(formattedRequests, gin.H{
			"id":           req.Id,
			"community_id": req.CommunityId,
			"user_id":      req.UserId,
			"status":       req.Status,
		})
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"join_requests": formattedRequests,
		"pagination": gin.H{
			"total_count":  len(formattedRequests),
			"current_page": 1,
			"per_page":     len(formattedRequests),
			"total_pages":  1,
		},
	})
}

func CreateChat(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("CreateChat: Missing userId in context")
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}
	log.Printf("CreateChat: Received request from user %v", userID)

	var req struct {
		Type         string   `json:"type"`
		Name         string   `json:"name"`
		Participants []string `json:"participants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateChat: JSON binding error: %v", err)
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request body: "+err.Error())
		return
	}
	log.Printf("CreateChat: Request data: type=%s, name=%s, participants=%v", req.Type, req.Name, req.Participants)

	if req.Type != "individual" && req.Type != "group" {
		log.Printf("CreateChat: Invalid chat type: %s", req.Type)
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid chat type, must be 'individual' or 'group'")
		return
	}

	if req.Type == "group" && req.Name == "" {
		log.Printf("CreateChat: Group chat missing name")
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Group chat name is required")
		return
	}

	if len(req.Participants) == 0 {
		log.Printf("CreateChat: No participants provided")
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "At least one participant is required")
		return
	}

	client := GetCommunityServiceClient()
	log.Printf("CreateChat: Got community service client")

	isGroup := req.Type == "group"
	name := req.Name
	log.Printf("CreateChat: Creating chat with isGroup=%v, name=%s", isGroup, name)

	chat, err := client.CreateChat(isGroup, name, req.Participants, userID.(string))
	if err != nil {
		log.Printf("CreateChat: Error from service: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to create chat: "+err.Error())
		return
	}
	log.Printf("CreateChat: Chat created successfully with ID %s", chat.ID)

	utils.SendSuccessResponse(c, 201, gin.H{
		"chat": chat,
	})
	log.Printf("CreateChat: Response sent with status 201")
}

func AddChatParticipant(c *gin.Context) {}

func RemoveChatParticipant(c *gin.Context) {}

func ListChats(c *gin.Context) {}

func ListChatParticipants(c *gin.Context) {}

func SendMessage(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("chatId")
	if chatID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send the message
	resp, err := CommunityClient.SendMessage(ctx, &communityProto.SendMessageRequest{
		ChatId:   chatID,
		SenderId: userID.(string),
		Content:  req.Content,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to send message: "+err.Error())
		return
	}

	// Format the response
	timestamp := time.Now().Unix()
	if resp.Message.SentAt != nil {
		timestamp = resp.Message.SentAt.AsTime().Unix()
	}

	utils.SendSuccessResponse(c, 201, gin.H{
		"message_id": resp.Message.Id, // In proto, there's no separate MessageId field
		"message": gin.H{
			"id":         resp.Message.Id,
			"chat_id":    resp.Message.ChatId,
			"sender_id":  resp.Message.SenderId,
			"content":    resp.Message.Content,
			"timestamp":  timestamp,
			"is_read":    !resp.Message.Unsent,
			"is_edited":  false,
			"is_deleted": resp.Message.DeletedForAll,
		},
	})
}

func DeleteMessage(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("chatId")
	if chatID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Message ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete the message
	_, err := CommunityClient.DeleteMessage(ctx, &communityProto.DeleteMessageRequest{
		MessageId: messageID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to delete message: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Message deleted successfully",
	})
}

func UnsendMessage(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Message ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if message belongs to user
	listResp, err := CommunityClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
		Limit: 1,
	})

	if err != nil || len(listResp.Messages) == 0 {
		utils.SendErrorResponse(c, 404, "NOT_FOUND", "Message not found")
		return
	}

	var message *communityProto.Message
	for _, msg := range listResp.Messages {
		if msg.Id == messageID {
			message = msg
			break
		}
	}

	if message == nil {
		utils.SendErrorResponse(c, 404, "NOT_FOUND", "Message not found")
		return
	}

	if message.SenderId != userID.(string) {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "You can only unsend your own messages")
		return
	}

	// Unsend the message
	_, err = CommunityClient.UnsendMessage(ctx, &communityProto.UnsendMessageRequest{
		MessageId: messageID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to unsend message: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Message unsent successfully",
	})
}

func ListMessages(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("chatId")
	if chatID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	limit := 20
	limitStr := c.DefaultQuery("limit", "20")
	if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
		limit = limitVal
	}

	offset := 0
	offsetStr := c.DefaultQuery("offset", "0")
	if offsetVal, err := strconv.Atoi(offsetStr); err == nil && offsetVal >= 0 {
		offset = offsetVal
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// List messages
	resp, err := CommunityClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
		ChatId: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list messages: "+err.Error())
		return
	}

	messages := make([]gin.H, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		timestamp := time.Now().Unix()
		if msg.SentAt != nil {
			timestamp = msg.SentAt.AsTime().Unix()
		}

		messages = append(messages, gin.H{
			"id":         msg.Id,
			"chat_id":    msg.ChatId,
			"sender_id":  msg.SenderId,
			"content":    msg.Content,
			"timestamp":  timestamp,
			"is_read":    !msg.Unsent,
			"is_edited":  false,
			"is_deleted": msg.DeletedForAll,
		})
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"messages": messages,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  len(messages),
		},
	})
}

func SearchMessages(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("chatId")
	if chatID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Search query is required")
		return
	}

	limit := 20
	limitStr := c.DefaultQuery("limit", "20")
	if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
		limit = limitVal
	}

	offset := 0
	offsetStr := c.DefaultQuery("offset", "0")
	if offsetVal, err := strconv.Atoi(offsetStr); err == nil && offsetVal >= 0 {
		offset = offsetVal
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Search messages
	resp, err := CommunityClient.SearchMessages(ctx, &communityProto.SearchMessagesRequest{
		ChatId: chatID,
		Query:  query,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to search messages: "+err.Error())
		return
	}

	messages := make([]gin.H, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		timestamp := time.Now().Unix()
		if msg.SentAt != nil {
			timestamp = msg.SentAt.AsTime().Unix()
		}

		messages = append(messages, gin.H{
			"id":         msg.Id,
			"chat_id":    msg.ChatId,
			"sender_id":  msg.SenderId,
			"content":    msg.Content,
			"timestamp":  timestamp,
			"is_read":    !msg.Unsent,
			"is_edited":  false,
			"is_deleted": msg.DeletedForAll,
		})
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"messages": messages,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  len(messages),
		},
	})
}

func GetDetailedChats(c *gin.Context) {}

func GetChatHistoryList(c *gin.Context) {}

func CheckMembershipStatus(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	memberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error checking membership status: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check membership status: "+err.Error())
		return
	}

	if memberResp.IsMember {
		utils.SendSuccessResponse(c, 200, gin.H{
			"status": "member",
		})
		return
	}

	pendingResp, err := CommunityClient.HasPendingJoinRequest(ctx, &communityProto.HasPendingJoinRequestRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error checking pending join request: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check join request status: "+err.Error())
		return
	}

	var status string
	if pendingResp.HasRequest {
		status = "pending"
	} else {
		status = "none"
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"status": status,
	})
}
