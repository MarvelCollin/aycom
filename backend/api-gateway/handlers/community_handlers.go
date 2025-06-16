package handlers

import (
	"aycom/backend/api-gateway/utils"
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"
	"context"
	"encoding/json"
	"log"
	"math"
	"regexp"
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
				// Even if creating the community request fails, we continue since the community itself was created
				// But we log a detailed error to help with troubleshooting
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
	offset := 0
	limit := 25

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	} else if pageStr := c.Query("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {

			if limitStr := c.Query("limit"); limitStr != "" {
				if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
					limit = limitInt
				}
			}
			offset = (pageInt - 1) * limit
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	query := c.Query("q")

	isApproved := true
	if isApprovedStr := c.Query("is_approved"); isApprovedStr != "" {
		isApproved = isApprovedStr == "true"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	var communities []*communityProto.Community
	var totalCount int32

	if query != "" {
		searchReq := &communityProto.SearchCommunitiesRequest{
			Query:      query,
			Offset:     int32(offset),
			Limit:      int32(limit),
			IsApproved: isApproved,
		}

		resp, err := CommunityClient.SearchCommunities(ctx, searchReq)
		if err != nil {
			log.Printf("Error calling SearchCommunities: %v", err)
			log.Printf("Falling back to ListCommunities")

			listResp, listErr := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
				Offset:     int32(offset),
				Limit:      int32(limit),
				IsApproved: isApproved,
			})

			if listErr != nil {
				log.Printf("Error calling ListCommunities: %v", listErr)
				utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list communities: "+err.Error())
				return
			}

			communities = listResp.Communities
			totalCount = listResp.TotalCount
		} else {
			communities = resp.Communities
			totalCount = resp.TotalCount
		}
	} else {

		resp, err := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
			Offset:     int32(offset),
			Limit:      int32(limit),
			IsApproved: isApproved,
		})

		if err != nil {
			log.Printf("Error calling ListCommunities: %v", err)
			utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list communities: "+err.Error())
			return
		}

		communities = resp.Communities
		totalCount = resp.TotalCount
	}

	result := make([]gin.H, 0, len(communities))
	for _, community := range communities {

		if community == nil {
			continue
		}

		categoryNames := make([]string, 0)
		if community.Categories != nil {
			for _, cat := range community.Categories {
				if cat != nil {
					categoryNames = append(categoryNames, cat.Name)
				}
			}
		}

		communityData := gin.H{
			"id":          community.Id,
			"name":        community.Name,
			"description": community.Description,
			"logo_url":    community.LogoUrl,
			"banner_url":  community.BannerUrl,
			"creator_id":  community.CreatorId,
			"is_approved": community.IsApproved,
			"categories":  categoryNames,
		}

		if community.CreatedAt != nil {
			communityData["created_at"] = community.CreatedAt.AsTime()
		}

		result = append(result, communityData)
	}

	totalPages := calculateTotalPages(int(totalCount), limit)
	currentPage := (offset / limit) + 1

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": result,
		"total_count": totalCount,
		"pagination": gin.H{
			"current_page": currentPage,
			"per_page":     limit,
			"total_pages":  totalPages,
			"total_count":  totalCount,
		},
		"limit_options": []int{25, 50, 100},
	})
}

func calculateTotalPages(totalCount, perPage int) int {
	if perPage <= 0 {
		return 1
	}
	if totalCount == 0 {
		return 0
	}
	return (totalCount + perPage - 1) / perPage
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
		req.Role = "member"
	}

	if CommunityClient == nil {
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {

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

	memberData := gin.H{
		"id":                  resp.Member.UserId,
		"user_id":             resp.Member.UserId,
		"username":            "user_" + resp.Member.UserId,
		"name":                "User " + resp.Member.UserId,
		"role":                resp.Member.Role,
		"joined_at":           joinedAt,
		"profile_picture_url": "",
	}

	if UserClient != nil {
		userCtx, userCancel := context.WithTimeout(context.Background(), 2*time.Second)
		userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
			UserId: resp.Member.UserId,
		})
		userCancel()

		if userErr == nil && userResp != nil && userResp.User != nil {
			user := userResp.User

			memberData = gin.H{
				"id":                  resp.Member.UserId,
				"user_id":             resp.Member.UserId,
				"username":            user.Username,
				"name":                user.Name,
				"role":                resp.Member.Role,
				"joined_at":           joinedAt,
				"profile_picture_url": user.ProfilePictureUrl,
				"is_verified":         user.IsVerified,
				"bio":                 user.Bio,
			}
		} else {
			log.Printf("Warning: Could not fetch user data for new member %s: %v", resp.Member.UserId, userErr)
		}
	} else {
		log.Printf("Warning: UserClient is nil, using placeholder data for new member %s", resp.Member.UserId)
	}

	utils.SendSuccessResponse(c, 201, memberData)
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

	selfRemoval := memberUserID == userID.(string)

	if !selfRemoval {
		isAdmin := false
		adminIDStr, adminExists := c.Get("isAdmin")
		if adminExists && adminIDStr.(bool) {
			isAdmin = true
		}

		if !isAdmin {

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

			memberData := gin.H{
				"id":                  member.UserId,
				"user_id":             member.UserId,
				"username":            "user_" + member.UserId,
				"name":                "User " + member.UserId,
				"role":                member.Role,
				"joined_at":           joinedAt,
				"profile_picture_url": "",
			}
			if UserClient != nil {
				userCtx, userCancel := context.WithTimeout(context.Background(), 2*time.Second)
				userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
					UserId: member.UserId,
				})
				userCancel()

				if userErr == nil && userResp != nil && userResp.User != nil {
					user := userResp.User

					memberData = gin.H{
						"id":                  member.UserId,
						"user_id":             member.UserId,
						"username":            user.Username,
						"name":                user.Name,
						"role":                member.Role,
						"joined_at":           joinedAt,
						"profile_picture_url": user.ProfilePictureUrl,
						"is_verified":         user.IsVerified,
						"bio":                 user.Bio,
					}
				} else {
					log.Printf("Warning: Could not fetch user data for member %s: %v", member.UserId, userErr)
				}
			} else {
				log.Printf("Warning: UserClient is nil, using placeholder data for member %s", member.UserId)
			}

			formattedMembers = append(formattedMembers, memberData)
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

	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {

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
		"username":            "user_" + resp.Member.UserId,
		"name":                "User " + resp.Member.UserId,
		"role":                resp.Member.Role,
		"joined_at":           joinedAt,
		"profile_picture_url": "",
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

	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {

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

	resp, err := CommunityClient.AddRule(ctx, &communityProto.AddRuleRequest{
		CommunityId: communityID,
		RuleText:    req.RuleText,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to add rule: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, 201, gin.H{
		"id":           resp.Rule.Id,
		"community_id": resp.Rule.CommunityId,
		"title":        "Community Rule",
		"description":  resp.Rule.RuleText,
		"order":        1,
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

	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {

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

	isAdmin := false

	if UserClient != nil {
		userCtx, userCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer userCancel()

		userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
			UserId: userID.(string),
		})

		if userErr == nil && userResp != nil && userResp.User != nil && userResp.User.IsAdmin {
			isAdmin = true
			log.Printf("User %s is a system admin, granting access to approve join request", userID.(string))
		}
	}

	if !isAdmin {

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

	isAdmin := false

	if UserClient != nil {
		userCtx, userCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer userCancel()

		userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
			UserId: userID.(string),
		})

		if userErr == nil && userResp != nil && userResp.User != nil && userResp.User.IsAdmin {
			isAdmin = true
			log.Printf("User %s is a system admin, granting access to reject join request", userID.(string))
		}
	}

	if !isAdmin {

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

	isAdmin := false
	adminIDStr, adminExists := c.Get("isAdmin")
	if adminExists && adminIDStr.(bool) {
		isAdmin = true
	}

	if !isAdmin {

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

	resp, err := CommunityClient.ListJoinRequests(ctx, &communityProto.ListJoinRequestsRequest{
		CommunityId: communityID,
	})

	if err != nil {
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list join requests: "+err.Error())
		return
	}
	formattedRequests := make([]gin.H, 0, len(resp.JoinRequests))
	for _, req := range resp.JoinRequests {

		requestData := gin.H{
			"id":                  req.Id,
			"community_id":        req.CommunityId,
			"user_id":             req.UserId,
			"status":              req.Status,
			"username":            "user_" + req.UserId,
			"name":                "User " + req.UserId,
			"profile_picture_url": "",
		}

		if UserClient != nil {
			userCtx, userCancel := context.WithTimeout(context.Background(), 2*time.Second)
			userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
				UserId: req.UserId,
			})
			userCancel()

			if userErr == nil && userResp != nil && userResp.User != nil {
				user := userResp.User

				requestData = gin.H{
					"id":                  req.Id,
					"community_id":        req.CommunityId,
					"user_id":             req.UserId,
					"status":              req.Status,
					"username":            user.Username,
					"name":                user.Name,
					"profile_picture_url": user.ProfilePictureUrl,
					"is_verified":         user.IsVerified,
					"bio":                 user.Bio,
				}
			} else {
				log.Printf("Warning: Could not fetch user data for join request %s: %v", req.UserId, userErr)
			}
		} else {
			log.Printf("Warning: UserClient is nil, using placeholder data for join request %s", req.UserId)
		}

		formattedRequests = append(formattedRequests, requestData)
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

func OldSearchCommunities(c *gin.Context) {
	query := c.Query("q")
	page := 1
	limit := 10

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

	var isApproved *bool
	if isApprovedStr := c.Query("is_approved"); isApprovedStr != "" {
		approved := isApprovedStr == "true"
		isApproved = &approved
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")

		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response *communityProto.ListCommunitiesResponse
	var err error

	if query == "" && len(categories) == 0 {
		listReq := &communityProto.ListCommunitiesRequest{
			Offset: int32((page - 1) * limit),
			Limit:  int32(limit),
		}

		if isApproved != nil {
			listReq.IsApproved = *isApproved
		}

		response, err = CommunityClient.ListCommunities(ctx, listReq)
	} else {

		searchReq := &communityProto.SearchCommunitiesRequest{
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
		}

		if isApproved != nil {
			searchReq.IsApproved = *isApproved
		}

		response, err = CommunityClient.SearchCommunities(ctx, searchReq)
	}

	if err != nil {
		log.Printf("Error in community search/list RPC call: %v", err)

		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	if response == nil || response.Communities == nil {
		log.Printf("Warning: Community search/list returned nil response or nil communities")

		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	communities := make([]map[string]interface{}, 0)
	for _, community := range response.Communities {

		if community == nil {
			log.Printf("Warning: nil community in response")
			continue
		}

		communityData := map[string]interface{}{
			"id":           community.Id,
			"name":         community.Name,
			"description":  community.Description,
			"logo_url":     community.LogoUrl,
			"banner_url":   community.BannerUrl,
			"creator_id":   community.CreatorId,
			"is_approved":  community.IsApproved,
			"member_count": 0,
		}

		if community.CreatedAt != nil {
			communityData["created_at"] = community.CreatedAt.AsTime()
		} else {
			communityData["created_at"] = time.Now()
		}

		if community.Categories != nil {
			categories := make([]string, 0)
			for _, category := range community.Categories {
				if category != nil {
					categories = append(categories, category.Name)
				}
			}
			communityData["categories"] = categories
		} else {
			communityData["categories"] = []string{}
		}

		communities = append(communities, communityData)
	}

	totalCount := int32(0)
	if response.TotalCount > 0 {
		totalCount = response.TotalCount
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": communities,
		"total_count": totalCount,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"total_pages":  totalPages,
		},
	})
}

func SearchCommunityByName(c *gin.Context) {
	rawQuery := c.Query("q")
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

	isApproved := true
	if isApprovedStr := c.Query("is_approved"); isApprovedStr != "" {
		isApproved = isApprovedStr == "true"
	}

	if rawQuery == "" {
		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9\\s]+")
	if err != nil {
		log.Printf("Regex compilation failed: %v", err)
		utils.SendErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to process search query")
		return
	}

	sanitizedQuery := reg.ReplaceAllString(rawQuery, " ")
	sanitizedQuery = strings.TrimSpace(sanitizedQuery)

	if sanitizedQuery == "" {
		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	if CommunityClient == nil {
		log.Printf("Error: CommunityClient is nil")
		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
		Offset:     int32((page - 1) * limit),
		Limit:      int32(100),
		IsApproved: isApproved,
	})

	if err != nil {
		log.Printf("Error calling ListCommunities: %v", err)
		utils.SendSuccessResponse(c, 200, gin.H{
			"communities": []gin.H{},
			"total_count": 0,
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"total_pages":  0,
			},
		})
		return
	}

	filtered := make([]*communityProto.Community, 0)

	queryTerms := strings.Fields(strings.ToLower(sanitizedQuery))

	for _, community := range resp.Communities {
		if community == nil {
			continue
		}

		communityName := strings.ToLower(community.Name)
		communityDesc := strings.ToLower(community.Description)

		matchFound := false
		for _, term := range queryTerms {
			if term == "" {
				continue
			}

			if strings.Contains(communityName, term) || strings.Contains(communityDesc, term) {
				matchFound = true
				break
			}
		}

		if matchFound {
			filtered = append(filtered, community)
		}
	}

	startIdx := 0
	endIdx := len(filtered)
	if endIdx > limit {
		endIdx = limit
	}

	pagedResults := filtered
	if startIdx < endIdx {
		pagedResults = filtered[startIdx:endIdx]
	} else {
		pagedResults = []*communityProto.Community{}
	}

	result := make([]gin.H, 0, len(pagedResults))
	for _, community := range pagedResults {
		categoryNames := make([]string, 0)
		if community.Categories != nil {
			for _, cat := range community.Categories {
				if cat != nil {
					categoryNames = append(categoryNames, cat.Name)
				}
			}
		}

		communityData := gin.H{
			"id":          community.Id,
			"name":        community.Name,
			"description": community.Description,
			"logo_url":    community.LogoUrl,
			"banner_url":  community.BannerUrl,
			"creator_id":  community.CreatorId,
			"is_approved": community.IsApproved,
			"categories":  categoryNames,
		}

		if community.CreatedAt != nil {
			communityData["created_at"] = community.CreatedAt.AsTime()
		}

		result = append(result, communityData)
	}

	totalCount := int32(len(filtered))
	totalPages := calculateTotalPages(int(totalCount), limit)

	utils.SendSuccessResponse(c, 200, gin.H{
		"communities": result,
		"total_count": totalCount,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_pages":  totalPages,
			"total_count":  totalCount,
		},
	})
}

func GetUserCommunities(c *gin.Context) {

	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	page := 1
	limit := 25
	filter := c.Query("filter")

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

	if filter == "joined" || filter == "pending" {

		status := filter
		resp, err = CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId:     userID.(string),
			Status:     status,
			Query:      query,
			Categories: categories,
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
		})
	} else if filter == "discover" {

		allCommunitiesResp, err := CommunityClient.SearchCommunities(ctx, &communityProto.SearchCommunitiesRequest{
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

		joinedResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID.(string),
			Status: "member",
			Limit:  1000,
		})

		if err != nil {
			log.Printf("Error fetching joined communities: %v", err)
		}

		pendingResp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID.(string),
			Status: "pending",
			Limit:  1000,
		})

		if err != nil {
			log.Printf("Error fetching pending communities: %v", err)
		}

		joinedCommunityMap := make(map[string]bool)
		pendingCommunityMap := make(map[string]bool)

		if joinedResp != nil && joinedResp.Communities != nil {
			for _, community := range joinedResp.Communities {
				joinedCommunityMap[community.Id] = true
			}
		}

		if pendingResp != nil && pendingResp.Communities != nil {
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

		resp = &communityProto.ListCommunitiesResponse{
			Communities: filteredCommunities,
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
	}

	if err != nil {
		log.Printf("Error fetching communities: %v", err)
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
				"id":          community.Id,
				"name":        community.Name,
				"description": community.Description,
				"logo_url":    community.LogoUrl,
				"banner_url":  community.BannerUrl,
				"creator_id":  community.CreatorId,
				"is_approved": community.IsApproved,
				"categories":  categoryNames,

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

func GetJoinedCommunities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.SendErrorResponse(c, 400, "INVALID_INPUT", "User ID is required")
		return
	}

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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var communities []gin.H
	var totalCount int32

	if CommunityClient != nil {
		resp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID,
			Status: "member",
			Offset: int32((page - 1) * limit),
			Limit:  int32(limit),
		})

		if err != nil {
			log.Printf("Error fetching joined communities: %v", err)
		} else if resp != nil {

			communities = make([]gin.H, 0, len(resp.Communities))
			if resp.Communities != nil {
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
						"id":          community.Id,
						"name":        community.Name,
						"description": community.Description,
						"logo_url":    community.LogoUrl,
						"banner_url":  community.BannerUrl,
						"creator_id":  community.CreatorId,
						"is_approved": community.IsApproved,
						"categories":  categoryNames,

						"member_count": 0,
					}

					if community.CreatedAt != nil {
						communityData["created_at"] = community.CreatedAt.AsTime()
					}

					communities = append(communities, communityData)
				}
			}

			totalCount = resp.TotalCount
		}
	}

	if communities == nil {
		communities = []gin.H{}
	}

	totalPages := calculateTotalPages(int(totalCount), limit)

	utils.SendDirectSuccessResponse(c, 200, gin.H{
		"communities": communities,
		"total":       totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func GetPendingCommunities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.SendErrorResponse(c, 400, "INVALID_INPUT", "User ID is required")
		return
	}

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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var communities []gin.H
	var totalCount int32

	if CommunityClient != nil {
		resp, err := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
			UserId: userID,
			Status: "pending",
			Offset: int32((page - 1) * limit),
			Limit:  int32(limit),
		})

		if err != nil {
			log.Printf("Error fetching pending communities: %v", err)
		} else if resp != nil {

			communities = make([]gin.H, 0, len(resp.Communities))
			if resp.Communities != nil {
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
						"id":          community.Id,
						"name":        community.Name,
						"description": community.Description,
						"logo_url":    community.LogoUrl,
						"banner_url":  community.BannerUrl,
						"creator_id":  community.CreatorId,
						"is_approved": community.IsApproved,
						"categories":  categoryNames,

						"member_count": 0,
					}

					if community.CreatedAt != nil {
						communityData["created_at"] = community.CreatedAt.AsTime()
					}

					communities = append(communities, communityData)
				}
			}

			totalCount = resp.TotalCount
		}
	}

	if communities == nil {
		communities = []gin.H{}
	}

	totalPages := calculateTotalPages(int(totalCount), limit)

	utils.SendDirectSuccessResponse(c, 200, gin.H{
		"communities": communities,
		"total":       totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func GetDiscoverCommunities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.SendErrorResponse(c, 400, "INVALID_INPUT", "User ID is required")
		return
	}

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

	isApproved := false
	if isApprovedStr := c.Query("is_approved"); isApprovedStr != "" {
		isApproved = isApprovedStr == "true"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var communities []gin.H
	var totalCount int32
	if CommunityClient != nil {

		allCommunitiesResp, err := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
			Offset:     int32((page - 1) * limit),
			Limit:      int32(limit),
			IsApproved: isApproved,
		})

		if err != nil {
			log.Printf("Error fetching all communities: %v", err)
		} else if allCommunitiesResp != nil && allCommunitiesResp.Communities != nil {

			joinedResp, joinedErr := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
				UserId: userID,
				Status: "member",

				Limit: 100,
			})

			pendingResp, pendingErr := CommunityClient.ListUserCommunities(ctx, &communityProto.ListUserCommunitiesRequest{
				UserId: userID,
				Status: "pending",

				Limit: 100,
			})

			joinedCommunityMap := make(map[string]bool)
			pendingCommunityMap := make(map[string]bool)

			if joinedErr == nil && joinedResp != nil && joinedResp.Communities != nil {
				for _, community := range joinedResp.Communities {
					joinedCommunityMap[community.Id] = true
				}
			}

			if pendingErr == nil && pendingResp != nil && pendingResp.Communities != nil {
				for _, community := range pendingResp.Communities {
					pendingCommunityMap[community.Id] = true
				}
			}

			communities = make([]gin.H, 0)
			for _, community := range allCommunitiesResp.Communities {
				if !joinedCommunityMap[community.Id] && !pendingCommunityMap[community.Id] {
					categoryNames := make([]string, 0)
					if community.Categories != nil {
						for _, cat := range community.Categories {
							if cat != nil {
								categoryNames = append(categoryNames, cat.Name)
							}
						}
					}

					communityData := gin.H{
						"id":          community.Id,
						"name":        community.Name,
						"description": community.Description,
						"logo_url":    community.LogoUrl,
						"banner_url":  community.BannerUrl,
						"creator_id":  community.CreatorId,
						"is_approved": community.IsApproved,
						"categories":  categoryNames,

						"member_count": 0,
					}

					if community.CreatedAt != nil {
						communityData["created_at"] = community.CreatedAt.AsTime()
					}

					communities = append(communities, communityData)
				}
			}

			totalCount = int32(len(communities))
		}
	}

	if communities == nil {
		communities = []gin.H{}
	}

	totalPages := calculateTotalPages(int(totalCount), limit)

	utils.SendDirectSuccessResponse(c, 200, gin.H{
		"communities": communities,
		"total":       totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
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
