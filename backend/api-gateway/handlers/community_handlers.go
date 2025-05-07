package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func CreateCommunity(c *gin.Context)  {}
func UpdateCommunity(c *gin.Context)  {}
func ApproveCommunity(c *gin.Context) {}
func DeleteCommunity(c *gin.Context)  {}
func GetCommunityByID(c *gin.Context) {}
func ListCommunities(c *gin.Context)  {}

func AddMember(c *gin.Context)        {}
func RemoveMember(c *gin.Context)     {}
func ListMembers(c *gin.Context)      {}
func UpdateMemberRole(c *gin.Context) {}

func AddRule(c *gin.Context)    {}
func RemoveRule(c *gin.Context) {}
func ListRules(c *gin.Context)  {}

func RequestToJoin(c *gin.Context)      {}
func ApproveJoinRequest(c *gin.Context) {}
func RejectJoinRequest(c *gin.Context)  {}
func ListJoinRequests(c *gin.Context)   {}

func CreateChat(c *gin.Context) {
	// Get user ID from context
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

func AddChatParticipant(c *gin.Context)    {}
func RemoveChatParticipant(c *gin.Context) {}
func ListChats(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	limit := 50 // Default limit
	offset := 0 // Default offset

	client := GetCommunityServiceClient()

	chats, err := client.GetChats(userID.(string), limit, offset)
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to fetch chats: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"chats":   chats,
	})
}
func ListChatParticipants(c *gin.Context) {}

func SendMessage(c *gin.Context)    {}
func DeleteMessage(c *gin.Context)  {}
func UnsendMessage(c *gin.Context)  {}
func ListMessages(c *gin.Context)   {}
func SearchMessages(c *gin.Context) {}
