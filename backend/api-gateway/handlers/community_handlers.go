package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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
func ListChatParticipants(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	chatID := c.Param("id")
	if chatID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Chat ID is required")
		return
	}

	client := GetCommunityServiceClient()

	// First check if user is a participant (has access to this chat)
	isParticipant, err := client.IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		log.Printf("Error checking if user is chat participant: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		SendErrorResponse(c, 403, "forbidden", "You don't have access to this chat")
		return
	}

	// Call the get participants API
	participants, err := client.GetChatParticipants(chatID)
	if err != nil {
		log.Printf("Error fetching chat participants: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to fetch participants: "+err.Error())
		return
	}

	// Enhance participant data with user information
	// In a real implementation, this would call the user service to get details
	// For now, we'll add some basic user information to at least have display names
	enhancedParticipants := make([]gin.H, len(participants))
	for i, p := range participants {
		// This is a simplified approach - in a real system we would fetch from a user service
		username := fmt.Sprintf("user%s", p[:4]) // Just using first few chars of ID for demo
		displayName := fmt.Sprintf("User %s", p[:4])

		enhancedParticipants[i] = gin.H{
			"id":           p,
			"user_id":      p,
			"username":     username,
			"display_name": displayName,
		}
	}

	c.JSON(200, gin.H{
		"success":      true,
		"participants": enhancedParticipants,
	})
}

func SendMessage(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	chatID := c.Param("id")
	if chatID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Chat ID is required")
		return
	}

	var request struct {
		Content   string `json:"content"`
		MessageID string `json:"message_id"` // Allow client to specify a temp ID
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		SendErrorResponse(c, 400, "invalid_request", "Invalid request body: "+err.Error())
		return
	}

	if request.Content == "" {
		SendErrorResponse(c, 400, "invalid_request", "Message content is required")
		return
	}

	// Get the community client and properly send the message to the backend service
	client := GetCommunityServiceClient()

	// This msgID will be returned by the backend service after saving to the database
	msgID, err := client.SendMessage(chatID, userID.(string), request.Content)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		SendErrorResponse(c, 500, "server_error", "Failed to send message: "+err.Error())
		return
	}

	// Get current time for timestamp
	currentTime := time.Now()
	timestamp := currentTime.Unix()

	// Generate simple user data for display purposes
	// In a real implementation, this would come from a user service
	shortenedId := userID.(string)
	if len(shortenedId) > 4 {
		shortenedId = shortenedId[:4]
	}

	username := fmt.Sprintf("user%s", shortenedId)
	displayName := fmt.Sprintf("User %s", shortenedId)

	// Create a properly formatted JSON response that matches frontend expectations
	response := gin.H{
		"success": true,
		"message": gin.H{
			"id":          msgID, // Use id for consistency
			"message_id":  msgID, // Also include message_id to match frontend
			"chat_id":     chatID,
			"sender_id":   userID,
			"user_id":     userID,
			"content":     request.Content,
			"timestamp":   timestamp,
			"original_id": request.MessageID, // Return the temp ID if provided
			"type":        "text",
			"is_edited":   false,
			"is_deleted":  false,
			"is_read":     false,
			"user": gin.H{
				"id":           userID,
				"username":     username,
				"display_name": displayName,
			},
		},
	}

	// Log the response for debugging
	responseBytes, _ := json.Marshal(response)
	log.Printf("Sending message response: %s", string(responseBytes))

	c.JSON(201, response)
}

func DeleteMessage(c *gin.Context) {}
func UnsendMessage(c *gin.Context) {}
func ListMessages(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	chatID := c.Param("id")
	if chatID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Chat ID is required")
		return
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	offset := 0 // Default offset

	client := GetCommunityServiceClient()

	// First check if user is a participant (by trying to get messages)
	// If the user isn't a participant, the GetMessages call will return an appropriate error
	messages, err := client.GetMessages(chatID, limit, offset)
	if err != nil {
		// Log who tried to access the chat
		log.Printf("User %v failed to fetch messages for chat %s: %v", userID, chatID, err)
		SendErrorResponse(c, 500, "server_error", "Failed to fetch messages: "+err.Error())
		return
	}

	// Format messages
	formattedMessages := make([]gin.H, 0, len(messages))
	for _, msg := range messages {
		// Format timestamp as Unix timestamp (seconds since epoch)
		timestamp := msg.Timestamp.Unix()

		msgObj := gin.H{
			"id":         msg.ID, // Include id
			"message_id": msg.ID, // Also include message_id for frontend compatibility
			"chat_id":    msg.ChatID,
			"sender_id":  msg.SenderID,
			"content":    msg.Content,
			"timestamp":  timestamp,
			"user_id":    msg.SenderID, // Add user_id for compatibility
			"is_edited":  msg.IsEdited,
			"is_deleted": msg.IsDeleted,
			"is_read":    msg.IsRead,
		}

		// Add user information when available
		if msg.SenderID != "" {
			// Generate a simple username and display name based on the sender ID
			// In a real implementation, this would fetch from a user service
			shortenedId := msg.SenderID
			if len(shortenedId) > 4 {
				shortenedId = shortenedId[:4]
			}

			username := fmt.Sprintf("user%s", shortenedId)
			displayName := fmt.Sprintf("User %s", shortenedId)

			msgObj["user"] = gin.H{
				"id":           msg.SenderID,
				"username":     username,
				"display_name": displayName,
			}
		}

		formattedMessages = append(formattedMessages, msgObj)
	}

	c.JSON(200, gin.H{
		"success":  true,
		"messages": formattedMessages,
	})
}
func SearchMessages(c *gin.Context) {}
