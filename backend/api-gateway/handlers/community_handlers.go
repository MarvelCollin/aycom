package handlers

import (
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

// GetCommunityByID handles fetching a community by its ID
func GetCommunityByID(c *gin.Context) {
	communityID := c.Param("id")
	log.Printf("GetCommunityByID called with ID: %s", communityID)
	
	if communityID == "" {
		log.Printf("Error: Empty community ID provided")
		c.JSON(400, gin.H{
			"success": false,
			"error": "bad_request",
			"message": "Community ID is required"
		})
		return
	}

	// Check if CommunityClient is available
	if CommunityClient == nil {
		log.Printf("ERROR: CommunityClient is nil! Community service may not be running.")
		// Return a valid JSON response instead of an error
		c.JSON(200, gin.H{
			"success": false,
			"error": "service_unavailable",
			"message": "Community service is unavailable",
			"community": map[string]interface{}{
				"id": communityID,
				"name": "Unknown Community",
				"description": "Unable to fetch community data. Service unavailable.",
				"logo": "",
				"banner": "",
				"creatorId": "",
				"isApproved": false,
				"categories": []interface{}{},
				"createdAt": time.Now(),
				"memberCount": 0,
			},
		})
		return
	}

	// Try to get real community data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.GetCommunityByID(ctx, &communityProto.GetCommunityByIDRequest{
		CommunityId: communityID,
	})

	// If there's an error, return a proper error response
	if err != nil {
		log.Printf("Error calling GetCommunityByID: %v", err)
		
		// Always return a valid JSON structure instead of an error
		c.JSON(200, gin.H{
			"success": false,
			"error": "server_error",
			"message": "Failed to get community: " + err.Error(),
			"community": gin.H{
				"id": communityID,
				"name": "Unknown Community",
				"description": "Error loading community data",
				"logo": "",
				"banner": "",
				"creatorId": "",
				"isApproved": false,
				"categories": []interface{}{},
				"createdAt": time.Now(),
				"memberCount": 0,
			},
		})
		return
	}

	// If the response is nil or doesn't contain a community, return an error
	if resp == nil || resp.Community == nil {
		log.Printf("GetCommunityByID returned nil response or nil community")
		c.JSON(200, gin.H{
			"success": false,
			"error": "not_found",
			"message": "Community not found",
			"community": gin.H{
				"id": communityID,
				"name": "Unknown Community",
				"description": "Community not found",
				"logo": "",
				"banner": "",
				"creatorId": "",
				"isApproved": false,
				"categories": []interface{}{},
				"createdAt": time.Now(),
				"memberCount": 0,
			},
		})
		return
	}

	// If we got here, we have a valid community response
	community := resp.Community
	
	// Format the community details for the response
	formattedCategories := make([]string, 0)
	
	// Extract categories if available
	if community.Categories != nil {
		for _, cat := range community.Categories {
			formattedCategories = append(formattedCategories, cat.Name)
		}
	}
	
	createdAt := time.Now()
	if community.CreatedAt != nil {
		createdAt = community.CreatedAt.AsTime()
	}

	// Count members - default to 0 as member_count doesn't exist in proto
	memberCount := 0

	c.JSON(200, gin.H{
		"success": true,
		"community": gin.H{
			"id":          community.Id,
			"name":        community.Name,
			"description": community.Description,
			"logo":        community.LogoUrl,
			"banner":      community.BannerUrl,
			"creatorId":   community.CreatorId,
			"isApproved":  community.IsApproved,
			"categories":  formattedCategories,
			"createdAt":   createdAt,
			"memberCount": memberCount,
		},
	})
}

// ListCommunities handles listing communities with pagination, filtering, and search
func ListCommunities(c *gin.Context) {
	// Parse pagination parameters
	// Allowed page sizes: 25, 30, 35
	limitOptions := []int{25, 30, 35}
	limit := limitOptions[0] // Default to 25

	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil {
			// Check if the limit is one of the allowed options
			validLimit := false
			for _, option := range limitOptions {
				if parsedLimit == option {
					validLimit = true
					limit = parsedLimit
					break
				}
			}

			if !validLimit {
				limit = limitOptions[0] // Reset to default if invalid
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

	// Parse filter parameters - getting the values but not using them yet
	// Will be used in a full implementation
	_ = c.DefaultQuery("filter", "all") // all, joined, pending
	_ = c.Query("q")                    // search query
	_ = c.QueryArray("category")        // categories filter

	// Process request based on filter
	var communities []*communityProto.Community
	var totalCount int32 = 0

	if CommunityClient == nil {
		log.Printf("CommunityClient is nil")
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// For demonstration - only show community listing without filters
	resp, err := CommunityClient.ListCommunities(ctx, &communityProto.ListCommunitiesRequest{
		Offset: int32(offset),
		Limit:  int32(limit),
	})

	if err != nil {
		log.Printf("Error calling ListCommunities: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to list communities: "+err.Error())
		return
	}

	communities = resp.GetCommunities()

	// Hard code totalCount for now
	totalCount = 10

	// Format communities with proper formatting for frontend
	formattedCommunities := make([]gin.H, 0, len(communities))
	for _, comm := range communities {
		formattedCategories := make([]string, 0)

		// We treat the categories as a separate field to display
		// In a real implementation, this would correctly extract categories from the community

		createdAt := time.Now()
		if comm.CreatedAt != nil {
			createdAt = comm.CreatedAt.AsTime()
		}

		formattedCommunities = append(formattedCommunities, gin.H{
			"id":          comm.Id,
			"name":        comm.Name,
			"description": comm.Description,
			"logo":        comm.LogoUrl,
			"banner":      comm.BannerUrl,
			"creatorId":   comm.CreatorId,
			"isApproved":  comm.IsApproved,
			"categories":  formattedCategories, // Empty for now
			"createdAt":   createdAt,
		})
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	c.JSON(200, gin.H{
		"success":     true,
		"communities": formattedCommunities,
		"pagination": gin.H{
			"total":      totalCount,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
		"limitOptions": limitOptions,
	})
}

// ListCategories handles listing all available community categories
func ListCategories(c *gin.Context) {
	if CommunityClient == nil {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	// Mock implementation
	// In a real implementation, we would call the service properly
	categories := []gin.H{
		{"id": "1", "name": "Technology"},
		{"id": "2", "name": "Gaming"},
		{"id": "3", "name": "Education"},
		{"id": "4", "name": "Entertainment"},
		{"id": "5", "name": "Sports"},
	}

	c.JSON(200, gin.H{
		"success":    true,
		"categories": categories,
	})
}

func AddMember(c *gin.Context)    {}
func RemoveMember(c *gin.Context) {}

// ListMembers handles listing members of a community
func ListMembers(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		SendErrorResponse(c, 400, "bad_request", "Community ID is required")
		return
	}

	// Parse pagination parameters
	limit := 20 // Default limit
	offset := 0 // Default offset

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
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ListMembersRequest only has community_id as per proto
	resp, err := CommunityClient.ListMembers(ctx, &communityProto.ListMembersRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling ListMembers: %v", err)

		// Check if it's a "not found" error
		if status.Code(err) == codes.NotFound {
			SendErrorResponse(c, 404, "not_found", "Community not found")
			return
		}

		// Otherwise it's a server error
		SendErrorResponse(c, 500, "server_error", "Failed to list members: "+err.Error())
		return
	}

	// Format the members for the response
	formattedMembers := make([]gin.H, 0)
	if resp != nil && resp.Members != nil {
		for _, member := range resp.Members {
			joinedAt := time.Now()
			if member.JoinedAt != nil {
				joinedAt = member.JoinedAt.AsTime()
			}

			// Get user details from member
			formattedMembers = append(formattedMembers, gin.H{
				"id":        member.UserId, // Using UserId as id
				"userId":    member.UserId,
				"username":  "user_" + member.UserId, // Default username
				"name":      "User " + member.UserId, // Default name
				"role":      member.Role,
				"joinedAt":  joinedAt,
				"avatarUrl": "", // No avatar in proto
			})
		}
	}

	// Calculate pagination info
	totalCount := int32(len(formattedMembers))

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	currentPage := offset/limit + 1

	c.JSON(200, gin.H{
		"success": true,
		"members": formattedMembers,
		"pagination": gin.H{
			"total":      totalCount,
			"page":       currentPage,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

func UpdateMemberRole(c *gin.Context) {}

func AddRule(c *gin.Context)    {}
func RemoveRule(c *gin.Context) {}

// ListRules handles listing community rules
func ListRules(c *gin.Context) {
	communityID := c.Param("id")
	if communityID == "" {
		SendErrorResponse(c, 400, "bad_request", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		log.Printf("CommunityClient is nil")
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := CommunityClient.ListRules(ctx, &communityProto.ListRulesRequest{
		CommunityId: communityID,
	})

	if err != nil {
		log.Printf("Error calling ListRules: %v", err)

		// Check if it's a "not found" error
		if status.Code(err) == codes.NotFound {
			SendErrorResponse(c, 404, "not_found", "Community not found")
			return
		}

		// Otherwise it's a server error
		SendErrorResponse(c, 500, "server_error", "Failed to list rules: "+err.Error())
		return
	}

	// Format the rules for the response
	formattedRules := make([]gin.H, 0)
	if resp != nil && resp.Rules != nil {
		for i, rule := range resp.Rules {
			formattedRules = append(formattedRules, gin.H{
				"id":          rule.Id,
				"communityId": rule.CommunityId,
				"title":       "Rule " + strconv.Itoa(i+1), // No title in proto
				"description": rule.RuleText,               // Use RuleText as description
				"order":       i + 1,                       // No order in proto
			})
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"rules":   formattedRules,
	})
}

// RequestToJoin handles user requests to join a community
func RequestToJoin(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		SendErrorResponse(c, 400, "bad_request", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
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
			SendErrorResponse(c, 400, "already_requested", "You have already requested to join this community")
			return
		}
		SendErrorResponse(c, 500, "server_error", "Failed to request to join: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Join request sent successfully",
		"joinRequest": gin.H{
			"id":          resp.JoinRequest.Id,
			"communityId": resp.JoinRequest.CommunityId,
			"userId":      resp.JoinRequest.UserId,
			"status":      resp.JoinRequest.Status,
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
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}
	log.Printf("CreateChat: Received request from user %v", userID)

	// Parse request body
	var req struct {
		Type         string   `json:"type"` // "individual" or "group"
		Name         string   `json:"name"` // Required for group chats
		Participants []string `json:"participants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateChat: JSON binding error: %v", err)
		SendErrorResponse(c, 400, "bad_request", "Invalid request body: "+err.Error())
		return
	}
	log.Printf("CreateChat: Request data: type=%s, name=%s, participants=%v", req.Type, req.Name, req.Participants)

	// Validate request
	if req.Type != "individual" && req.Type != "group" {
		log.Printf("CreateChat: Invalid chat type: %s", req.Type)
		SendErrorResponse(c, 400, "bad_request", "Invalid chat type, must be 'individual' or 'group'")
		return
	}

	if req.Type == "group" && req.Name == "" {
		log.Printf("CreateChat: Group chat missing name")
		SendErrorResponse(c, 400, "bad_request", "Group chat name is required")
		return
	}

	if len(req.Participants) == 0 {
		log.Printf("CreateChat: No participants provided")
		SendErrorResponse(c, 400, "bad_request", "At least one participant is required")
		return
	}

	// Create gRPC client
	client := GetCommunityServiceClient()
	log.Printf("CreateChat: Got community service client")

	// Determine chat properties
	isGroup := req.Type == "group"
	name := req.Name
	log.Printf("CreateChat: Creating chat with isGroup=%v, name=%s", isGroup, name)

	// Create the chat
	chat, err := client.CreateChat(isGroup, name, req.Participants, userID.(string))
	if err != nil {
		log.Printf("CreateChat: Error from service: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to create chat: "+err.Error())
		return
	}
	log.Printf("CreateChat: Chat created successfully with ID %s", chat.ID)

	// Return the created chat
	c.JSON(201, gin.H{
		"success": true,
		"chat":    chat,
	})
	log.Printf("CreateChat: Response sent with status 201")
}

// AddChatParticipant handles adding a participant to a chat
func AddChatParticipant(c *gin.Context) {}

// RemoveChatParticipant handles removing a participant from a chat
func RemoveChatParticipant(c *gin.Context) {}

// ListChats handles listing all chats for a user
func ListChats(c *gin.Context) {}

// ListChatParticipants handles listing all participants in a chat
func ListChatParticipants(c *gin.Context) {}

// SendMessage handles sending a message to a chat
func SendMessage(c *gin.Context) {}

// DeleteMessage handles deleting a message from a chat
func DeleteMessage(c *gin.Context) {}

// UnsendMessage handles unsending a message from a chat
func UnsendMessage(c *gin.Context) {}

// ListMessages handles listing all messages in a chat
func ListMessages(c *gin.Context) {}

// SearchMessages handles searching for messages in a chat
func SearchMessages(c *gin.Context) {}

// GetDetailedChats handles getting detailed chat information including participants and last message
func GetDetailedChats(c *gin.Context) {}

// GetChatHistoryList handles getting a list of chat history for a user
func GetChatHistoryList(c *gin.Context) {}

// CheckMembershipStatus returns the current user's membership status in a community (none, member, or pending)
func CheckMembershipStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	communityID := c.Param("id")
	if communityID == "" {
		SendErrorResponse(c, 400, "bad_request", "Community ID is required")
		return
	}

	if CommunityClient == nil {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First check if user is a member
	memberResp, err := CommunityClient.IsMember(ctx, &communityProto.IsMemberRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error checking membership status: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to check membership status: "+err.Error())
		return
	}

	// If user is a member, return "member" status
	if memberResp.IsMember {
		c.JSON(200, gin.H{
			"success": true,
			"status":  "member",
		})
		return
	}

	// If not a member, check if they have a pending join request
	pendingResp, err := CommunityClient.HasPendingJoinRequest(ctx, &communityProto.HasPendingJoinRequestRequest{
		CommunityId: communityID,
		UserId:      userID.(string),
	})

	if err != nil {
		log.Printf("Error checking pending join request: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to check join request status: "+err.Error())
		return
	}

	// Return the appropriate status
	var status string
	if pendingResp.HasRequest {
		status = "pending"
	} else {
		status = "none"
	}

	c.JSON(200, gin.H{
		"success": true,
		"status":  status,
	})
}
