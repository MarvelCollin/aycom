package handlers

import (
	"aycom/backend/api-gateway/utils"
	communityProto "aycom/backend/proto/community"
	"context"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateCommunity(c *gin.Context)  {}
func UpdateCommunity(c *gin.Context)  {}
func ApproveCommunity(c *gin.Context) {}
func DeleteCommunity(c *gin.Context)  {}

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

	totalCount = 10

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

	categories := []gin.H{
		{"id": "1", "name": "Technology"},
		{"id": "2", "name": "Gaming"},
		{"id": "3", "name": "Education"},
		{"id": "4", "name": "Entertainment"},
		{"id": "5", "name": "Sports"},
	}

	utils.SendSuccessResponse(c, 200, gin.H{
		"categories": categories,
	})
}

func AddMember(c *gin.Context)    {}
func RemoveMember(c *gin.Context) {}

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

func UpdateMemberRole(c *gin.Context) {}

func AddRule(c *gin.Context)    {}
func RemoveRule(c *gin.Context) {}

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

func ApproveJoinRequest(c *gin.Context) {}
func RejectJoinRequest(c *gin.Context)  {}
func ListJoinRequests(c *gin.Context)   {}

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

func SendMessage(c *gin.Context) {}

func DeleteMessage(c *gin.Context) {}

func UnsendMessage(c *gin.Context) {}

func ListMessages(c *gin.Context) {}

func SearchMessages(c *gin.Context) {}

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
