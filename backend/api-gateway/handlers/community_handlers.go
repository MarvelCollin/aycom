package handlers

import (
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

	"aycom/backend/api-gateway/utils"
)

func calculateTotalPages(total, perPage int) int {
	if total <= 0 || perPage <= 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}

func CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	contentType := c.GetHeader("Content-Type")
	log.Printf("DEBUG: Content-Type header: %s", contentType)

	var name, description, logoURL, bannerURL, rules string
	var categories []string

	if strings.HasPrefix(contentType, "multipart/form-data") {

		log.Printf("DEBUG: Handling multipart form data")
		name = c.PostForm("name")
		description = c.PostForm("description")
		rules = c.PostForm("rules")
		log.Printf("DEBUG: Received form values - name: %s, description: %s, rules: %s", name, description, rules)

		categoriesJSON := c.PostForm("categories")
		log.Printf("DEBUG: Categories JSON: %s", categoriesJSON)

		if categoriesJSON != "" {
			if err := json.Unmarshal([]byte(categoriesJSON), &categories); err != nil {
				log.Printf("ERROR: Invalid categories format: %v", err)
				utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid categories format")
				return
			}
			log.Printf("DEBUG: Parsed categories: %v", categories)
		}

		logoFile, err := c.FormFile("icon")
		if err != nil {
			log.Printf("ERROR: No logo file: %v", err)
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Logo file is required")
			return
		}
		log.Printf("DEBUG: Logo file received: %s, size: %d", logoFile.Filename, logoFile.Size)

		bannerFile, err := c.FormFile("banner")
		if err != nil {
			log.Printf("ERROR: No banner file: %v", err)
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Banner file is required")
			return
		}
		log.Printf("DEBUG: Banner file received: %s, size: %d", bannerFile.Filename, bannerFile.Size)

		logoFileOpen, err := logoFile.Open()
		if err != nil {
			log.Printf("ERROR: Failed to open logo file: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to process logo file")
			return
		}
		defer logoFileOpen.Close()

		bannerFileOpen, err := bannerFile.Open()
		if err != nil {
			log.Printf("ERROR: Failed to open banner file: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to process banner file")
			return
		}
		defer bannerFileOpen.Close()

		log.Printf("DEBUG: Attempting to upload logo to Supabase bucket: media, folder: communities/logos")
		logoURL, err = utils.UploadFile(logoFileOpen, logoFile.Filename, "media", "communities/logos")
		if err != nil {
			log.Printf("ERROR: Failed to upload logo to Supabase: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to upload logo")
			return
		}
		log.Printf("DEBUG: Successfully uploaded logo, URL: %s", logoURL)

		log.Printf("DEBUG: Attempting to upload banner to Supabase bucket: media, folder: communities/banners")
		bannerURL, err = utils.UploadFile(bannerFileOpen, bannerFile.Filename, "media", "communities/banners")
		if err != nil {
			log.Printf("ERROR: Failed to upload banner to Supabase: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to upload banner")
			return
		}
		log.Printf("DEBUG: Successfully uploaded banner, URL: %s", bannerURL)

		if name == "" || description == "" || rules == "" || len(categories) == 0 {
			log.Printf("CreateCommunity: Missing required fields")
			utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Missing required fields")
			return
		}
	} else {

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

	protoCategories := make([]*communityProto.Category, len(categories))
	for i, categoryName := range categories {
		protoCategories[i] = &communityProto.Category{
			Name: categoryName,
		}
	}

	community := &communityProto.Community{
		Name:        name,
		Description: description,
		LogoUrl:     logoURL,
		BannerUrl:   bannerURL,
		CreatorId:   userID.(string),
		IsApproved:  false,
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

	if resp != nil && resp.Community != nil {
		_, err = CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
			CommunityId: resp.Community.Id,
			RuleText:    rules,
		})

		if err != nil {
			log.Printf("Error adding rules to community: %v", err)
		}
	}

	if resp != nil && resp.Community != nil {
		_, err = CommunityClient.AddMember(ctx, &communityProto.AddMemberRequest{
			CommunityId: resp.Community.Id,
			UserId:      userID.(string),
			Role:        "admin",
		})

		if err != nil {
			log.Printf("Error adding creator as member: %v", err)
		}

		if UserClient != nil {
			userCtx, userCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer userCancel()

			log.Printf("Creating community request in user service for community ID: %s", resp.Community.Id)
			createReqResp, err := UserClient.CreateCommunityRequest(userCtx, &userProto.CreateCommunityRequestRequest{
				CommunityId: resp.Community.Id,
				UserId:      userID.(string),
				Name:        name,
				Description: description,
			})

			if err != nil {
				log.Printf("ERROR: Failed to create community request in user service: %v", err)

				log.Printf("SYNC WARNING: Community with ID %s exists in community service but not in community_requests table", resp.Community.Id)
			} else {
				log.Printf("Successfully created community request in user service for community ID: %s, request ID: %s",
					resp.Community.Id, createReqResp.Request.Id)
			}
		} else {
			log.Printf("ERROR: UserClient is nil! Could not create community request in user service")
			log.Printf("SYNC WARNING: Community with ID %s exists in community service but not in community_requests table", resp.Community.Id)
		}
	}

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
		"member_count": 1,
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

	getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error getting community for update: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
		return
	}

	if getCommunityResp.Community.CreatorId != userID.(string) {

		isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
			CommunityId: communityID,
			UserId:      userID.(string),
		})

		if err != nil || !isMemberResp.IsMember {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community creator or admin can update the community")
			return
		}

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

	categories := make([]*communityProto.Category, len(req.Categories))
	for i, categoryName := range req.Categories {
		categories[i] = &communityProto.Category{
			Name: categoryName,
		}
	}

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

	if req.Rules != "" {

		rulesResp, err := CommunityClient.ListRules(ctx, &communityProto.ListRulesRequest{
			CommunityId: communityID,
		})

		if err == nil && rulesResp.Rules != nil {
			for _, rule := range rulesResp.Rules {
				_, _ = CommunityClient.RemoveRule(ctx, &communityProto.RemoveRuleRequest{
					RuleId: rule.Id,
				})
			}
		}

		_, err = CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
			CommunityId: communityID,
			RuleText:    req.Rules,
		})

		if err != nil {
			log.Printf("Error updating rules: %v", err)

		}
	}

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

	getCommunityResp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error getting community for deletion: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to get community: "+err.Error())
		return
	}

	if getCommunityResp.Community.CreatorId != userID.(string) {

		isAdmin := false
		adminIDStr, adminExists := c.Get("isAdmin")
		if adminExists && adminIDStr.(bool) {
			isAdmin = true
		}

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

	memberCount := int64(0)

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

	var userID string
	tokenUserID, exists := c.Get("userId")

	if exists {
		userID = tokenUserID.(string)
		log.Printf("Authenticated user ID: %s", userID)
	} else {

		if paramUserID := c.Param("userId"); paramUserID != "" {
			userID = paramUserID
			log.Printf("Using URL parameter user ID: %s", userID)
		} else if queryUserID := c.Query("userId"); queryUserID != "" {
			userID = queryUserID
			log.Printf("Using query parameter user ID: %s", userID)
		}
	}

	page := 1
	limit := 25

	filter := c.Query("filter")
	log.Printf("Processing community list request with filter: %s, userID: %s", filter, userID)

	if pageStr := c.Query("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	query := c.Query("q")
	var categories []string
	if categoriesParam := c.QueryArray("category"); len(categoriesParam) > 0 {
		categories = categoriesParam
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp *communityProto.ListCommunitiesResponse
	var err error

	if filter == "joined" && userID != "" {

		userCommunitiesResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId:     userID,
			Status:     "member",
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
		})

		if err != nil {
			log.Printf("Error fetching joined communities: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch joined communities")
			return
		}

		resp = &communityProto.ListCommunitiesResponse{
			Communities: userCommunitiesResp.Communities,
			TotalCount:  userCommunitiesResp.TotalCount,
		}
	} else if filter == "pending" && userID != "" {

		userCommunitiesResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId:     userID,
			Status:     "pending",
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
		})

		if err != nil {
			log.Printf("Error fetching pending communities: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch pending communities")
			return
		}

		resp = &communityProto.ListCommunitiesResponse{
			Communities: userCommunitiesResp.Communities,
			TotalCount:  userCommunitiesResp.TotalCount,
		}
	} else if filter == "discover" && userID != "" {

		allCommunitiesResp, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit * 2),
			IsApproved: true,
		})

		if err != nil {
			log.Printf("Error fetching all communities: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch communities")
			return
		}

		joinedResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID,
			Status: "member",
			Limit:  1000,
		})

		joinedCommunityMap := make(map[string]bool)
		if err == nil && joinedResp != nil && joinedResp.Communities != nil {
			for _, community := range joinedResp.Communities {
				joinedCommunityMap[community.Id] = true
			}
		}

		pendingResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID,
			Status: "pending",
			Limit:  1000,
		})

		pendingCommunityMap := make(map[string]bool)
		if err == nil && pendingResp != nil && pendingResp.Communities != nil {
			for _, community := range pendingResp.Communities {
				pendingCommunityMap[community.Id] = true
			}
		}

		var filteredCommunities []*communityProto.Community
		for _, community := range allCommunitiesResp.Communities {
			if !joinedCommunityMap[community.Id] && !pendingCommunityMap[community.Id] {
				filteredCommunities = append(filteredCommunities, community)
			}
		}

		endIdx := limit
		if endIdx > len(filteredCommunities) {
			endIdx = len(filteredCommunities)
		}

		var pagedCommunities []*communityProto.Community
		if len(filteredCommunities) > 0 {
			pagedCommunities = filteredCommunities[:endIdx]
		}

		resp = &communityProto.ListCommunitiesResponse{
			Communities: pagedCommunities,
			TotalCount:  int32(len(filteredCommunities)),
		}
	} else {

		resp, err = CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
			IsApproved: true,
		})

		if err != nil {
			log.Printf("Error fetching communities: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch communities")
			return
		}
	}

	communitiesResult := make([]gin.H, 0)
	if resp != nil && resp.Communities != nil {
		for _, community := range resp.Communities {
			categoryNames := make([]string, 0)
			if community.Categories != nil {
				for _, cat := range community.Categories {
					if cat != nil {
						categoryNames = append(categoryNames, cat.Name)
					}
				}
			}

			communityData := gin.H{
				"id":           community.Id,
				"name":         community.Name,
				"description":  community.Description,
				"logo_url":     community.LogoUrl,
				"banner_url":   community.BannerUrl,
				"creator_id":   community.CreatorId,
				"is_approved":  community.IsApproved,
				"categories":   categoryNames,
				"member_count": 0,
			}

			if community.CreatedAt != nil {
				communityData["created_at"] = community.CreatedAt.AsTime()
			}

			communitiesResult = append(communitiesResult, communityData)
		}
	}

	totalCount := 0
	if resp != nil {
		totalCount = int(resp.TotalCount)
	}
	totalPages := calculateTotalPages(totalCount, limit)
	currentPage := page

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": communitiesResult,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": currentPage,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
		"limit_options": []int{25, 30, 35},
	})
}

func GetDiscoverCommunities(c *gin.Context) {

	page := 1
	limit := 25

	if pageStr := c.Query("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	query := c.Query("q")
	var categories []string
	if categoriesParam := c.QueryArray("category"); len(categoriesParam) > 0 {
		categories = categoriesParam
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Getting discover communities")

	resp, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
		Query:      query,
		Categories: categories,
		Offset:     int32((page - 1) * limit),
		Limit:      int32(limit),
		IsApproved: true,
	})

	if err != nil {
		log.Printf("Error fetching all communities: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch communities")
		return
	}

	communitiesResult := make([]gin.H, 0)
	if resp != nil && resp.Communities != nil {
		for _, community := range resp.Communities {
			categoryNames := make([]string, 0)
			if community.Categories != nil {
				for _, cat := range community.Categories {
					if cat != nil {
						categoryNames = append(categoryNames, cat.Name)
					}
				}
			}

			communityData := gin.H{
				"id":           community.Id,
				"name":         community.Name,
				"description":  community.Description,
				"logo_url":     community.LogoUrl,
				"banner_url":   community.BannerUrl,
				"creator_id":   community.CreatorId,
				"is_approved":  community.IsApproved,
				"categories":   categoryNames,
				"member_count": 0,
			}

			if community.CreatedAt != nil {
				communityData["created_at"] = community.CreatedAt.AsTime()
			}

			communitiesResult = append(communitiesResult, communityData)
		}
	}

	totalCount := 0
	if resp != nil {
		totalCount = int(resp.TotalCount)
	}
	totalPages := calculateTotalPages(totalCount, limit)
	currentPage := page

	result := gin.H{
		"communities": communitiesResult,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": currentPage,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
		"limit_options": []int{25, 30, 35},
	}

	utils.SendSuccessResponse(c, 200, result)
}

func GetTopCommunityMembers(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	c.Params = append(c.Params, gin.Param{Key: "id", Value: communityID})
	ListMembers(c)
}

func GetCommunityThreadsByLikes(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"threads":     []gin.H{},
		"total_count": 0,
		"pagination": gin.H{
			"current_page": 1,
			"per_page":     10,
			"total_pages":  0,
		},
	})
}

func GetCommunityThreadsByDate(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"threads":     []gin.H{},
		"total_count": 0,
		"pagination": gin.H{
			"current_page": 1,
			"per_page":     10,
			"total_pages":  0,
		},
	})
}

func GetCommunityMediaThreads(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"threads":     []gin.H{},
		"total_count": 0,
		"pagination": gin.H{
			"current_page": 1,
			"per_page":     10,
			"total_pages":  0,
		},
	})
}

func OldSearchCommunities(c *gin.Context) {
	query := c.Query("q")
	page := 1
	limit := 25

	if pageStr := c.Query("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	var categories []string
	if categoriesParam := c.QueryArray("category"); len(categoriesParam) > 0 {
		categories = categoriesParam
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
		Query:      query,
		Categories: categories,
		Offset:     int32((page - 1) * limit),
		Limit:      int32(limit),
		IsApproved: true,
	})

	if err != nil {
		log.Printf("Error searching communities: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to search communities")
		return
	}

	communitiesResult := make([]gin.H, 0)
	if resp != nil && resp.Communities != nil {
		for _, community := range resp.Communities {
			categoryNames := make([]string, 0)
			if community.Categories != nil {
				for _, cat := range community.Categories {
					if cat != nil {
						categoryNames = append(categoryNames, cat.Name)
					}
				}
			}

			communityData := gin.H{
				"id":           community.Id,
				"name":         community.Name,
				"description":  community.Description,
				"logo_url":     community.LogoUrl,
				"banner_url":   community.BannerUrl,
				"creator_id":   community.CreatorId,
				"is_approved":  community.IsApproved,
				"categories":   categoryNames,
				"member_count": 0,
			}

			if community.CreatedAt != nil {
				communityData["created_at"] = community.CreatedAt.AsTime()
			}

			communitiesResult = append(communitiesResult, communityData)
		}
	}

	totalCount := 0
	if resp != nil {
		totalCount = int(resp.TotalCount)
	}
	totalPages := calculateTotalPages(totalCount, limit)
	currentPage := page

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": communitiesResult,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": currentPage,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
		"limit_options": []int{25, 30, 35},
	})
}

func ListCategories(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "community_categories"

	
	var cachedResponse gin.H
	if err := utils.GetCache(ctx, cacheKey, &cachedResponse); err == nil {
		c.Header("X-Cache", "HIT")
		utils.SendSuccessResponse(c, 200, cachedResponse)
		return
	}

	
	c.Header("X-Cache", "MISS")

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListCategories(ctx, &communityProto.ListCategoriesRequest{})
	if err != nil {
		log.Printf("Error listing categories: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list categories")
		return
	}

	categories := make([]gin.H, 0)
	if resp != nil && resp.Categories != nil {
		for _, cat := range resp.Categories {
			categoryData := gin.H{
				"id":   cat.Id,
				"name": cat.Name,
			}
			categories = append(categories, categoryData)
		}
	}

	response := gin.H{
		"categories": categories,
	}

	
	_ = utils.SetCache(context.Background(), cacheKey, response, 12*time.Hour)

	utils.SendSuccessResponse(c, 200, response)
}

func ListMembers(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})
	if err != nil {
		log.Printf("Error listing members: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list members")
		return
	}

	members := make([]gin.H, 0)
	if resp != nil && resp.Members != nil {
		for _, member := range resp.Members {
			memberData := gin.H{
				"user_id":   member.UserId,
				"role":      member.Role,
				"joined_at": member.JoinedAt.AsTime(),
			}
			members = append(members, memberData)
		}
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"members": members,
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
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil || !isMemberResp.IsMember {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community members can add new members")
		return
	}

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can add new members")
		return
	}

	resp, err := CommunityClient.AddMember(ctx, &communityProto.AddMemberRequest{
		CommunityId: communityID,
		UserId:      req.UserID,
		Role:        req.Role,
	})

	if err != nil {
		log.Printf("Error adding member: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to add member")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Member added successfully",
		"member": gin.H{
			"user_id":   resp.Member.UserId,
			"role":      resp.Member.Role,
			"joined_at": resp.Member.JoinedAt.AsTime(),
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

	memberID := c.Param("userId")
	if memberID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Member ID is required")
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil || !isMemberResp.IsMember {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community members can update roles")
		return
	}

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can update roles")
		return
	}

	resp, err := CommunityClient.UpdateMemberRole(ctx, &communityProto.UpdateMemberRoleRequest{
		CommunityId: communityID,
		UserId:      memberID,
		Role:        req.Role,
	})

	if err != nil {
		log.Printf("Error updating member role: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to update member role")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Member role updated successfully",
		"member": gin.H{
			"user_id": resp.Member.UserId,
			"role":    resp.Member.Role,
		},
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

	memberID := c.Param("userId")
	if memberID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Member ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if userID.(string) != memberID {

		isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
			CommunityId: communityID,
			UserId:      userID.(string),
		})

		if err != nil || !isMemberResp.IsMember {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community members can remove members")
			return
		}

		membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
			CommunityId: communityID,
		})

		isAdmin := false
		if err == nil && membersResp.Members != nil {
			for _, member := range membersResp.Members {
				if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
					isAdmin = true
					break
				}
			}
		}

		if !isAdmin {
			utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can remove members")
			return
		}
	}

	_, err := CommunityClient.RemoveMember(ctx, &communityProto.RemoveMemberRequest{
		CommunityId: communityID,
		UserId:      memberID,
	})

	if err != nil {
		log.Printf("Error removing member: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to remove member")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Member removed successfully",
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can add rules")
		return
	}

	resp, err := CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
		CommunityId: communityID,
		RuleText:    req.RuleText,
	})

	if err != nil {
		log.Printf("Error adding rule: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to add rule")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Rule added successfully",
		"rule": gin.H{
			"id":         resp.Rule.Id,
			"rule_text":  resp.Rule.RuleText,
			"created_at": resp.Rule.CreatedAt.AsTime(),
		},
	})
}

func ListRules(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListRules(ctx, &communityProto.ListRulesRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error listing rules: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list rules")
		return
	}

	rules := make([]gin.H, 0)
	if resp != nil && resp.Rules != nil {
		for _, rule := range resp.Rules {
			ruleData := gin.H{
				"id":         rule.Id,
				"rule_text":  rule.RuleText,
				"created_at": rule.CreatedAt.AsTime(),
			}
			rules = append(rules, ruleData)
		}
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"rules": rules,
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can remove rules")
		return
	}

	_, err = CommunityClient.RemoveRule(ctx, &communityProto.RemoveRuleRequest{
		RuleId: ruleID,
	})

	if err != nil {
		log.Printf("Error removing rule: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to remove rule")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Rule removed successfully",
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err == nil && isMemberResp.IsMember {
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "User is already a member of this community")
		return
	}

	resp, err := CommunityClient.RequestToJoin(ctx, &communityProto.RequestToJoinRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error requesting to join: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to request to join")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request sent successfully",
		"request": gin.H{
			"id":         resp.JoinRequest.Id,
			"user_id":    resp.JoinRequest.UserId,
			"created_at": resp.JoinRequest.CreatedAt.AsTime(),
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can view join requests")
		return
	}

	resp, err := CommunityClient.ListJoinRequests(ctx, &communityProto.ListJoinRequestsRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error listing join requests: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list join requests")
		return
	}

	joinRequests := make([]gin.H, 0)
	if resp != nil && resp.JoinRequests != nil {
		for _, request := range resp.JoinRequests {
			requestData := gin.H{
				"id":         request.Id,
				"user_id":    request.UserId,
				"created_at": request.CreatedAt.AsTime(),
			}
			joinRequests = append(joinRequests, requestData)
		}
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"join_requests": joinRequests,
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can approve join requests")
		return
	}

	_, err = CommunityClient.ApproveJoinRequest(ctx, &communityProto.ApproveJoinRequestRequest{
		JoinRequestId: requestID,
	})

	if err != nil {
		log.Printf("Error approving join request: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to approve join request")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request approved successfully",
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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	isAdmin := false
	if err == nil && membersResp.Members != nil {
		for _, member := range membersResp.Members {
			if member.UserId == userID.(string) && (member.Role == "admin" || member.Role == "creator") {
				isAdmin = true
				break
			}
		}
	}

	if !isAdmin {
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "Only community admins can reject join requests")
		return
	}

	_, err = CommunityClient.RejectJoinRequest(ctx, &communityProto.RejectJoinRequestRequest{
		JoinRequestId: requestID,
	})

	if err != nil {
		log.Printf("Error rejecting join request: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to reject join request")
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"message": "Join request rejected successfully",
	})
}

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
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isMemberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error checking membership: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to check membership status")
		return
	}

	if isMemberResp.IsMember {

		membersResp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
			CommunityId: communityID,
		})

		userRole := "member"
		if err == nil && membersResp.Members != nil {
			for _, member := range membersResp.Members {
				if member.UserId == userID.(string) {
					userRole = member.Role
					break
				}
			}
		}

		utils.SendSuccessResponse(c, 200, gin.H{
			"status":    "member",
			"is_member": true,
			"user_role": userRole,
		})
		return
	}

	pendingResp, err := CommunityClient.HasPendingJoinRequest(ctx, &communityProto.HasPendingJoinRequestRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err == nil && pendingResp.GetHasRequest() {
		utils.SendSuccessResponse(c, 200, gin.H{
			"status":    "pending",
			"is_member": false,
		})
		return
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"status":    "none",
		"is_member": false,
	})
}
