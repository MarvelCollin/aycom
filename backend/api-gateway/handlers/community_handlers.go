package handlers

import (
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"
	"context"
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

func AddChatParticipant(c *gin.Context) {
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

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SendErrorResponse(c, 400, "bad_request", "Invalid request body: "+err.Error())
		return
	}

	if req.UserID == "" {
		SendErrorResponse(c, 400, "bad_request", "User ID is required")
		return
	}

	client := GetCommunityServiceClient()

	// Check if user is participant in the chat
	isParticipant, err := client.IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		SendErrorResponse(c, 403, "forbidden", "You don't have access to this chat")
		return
	}

	// Use the gRPC client directly as our service client doesn't expose this method
	if CommunityClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = CommunityClient.AddChatParticipant(ctx, &communityProto.AddChatParticipantRequest{
			ChatId: chatID,
			UserId: req.UserID,
		})
		if err != nil {
			SendErrorResponse(c, 500, "server_error", "Failed to add participant: "+err.Error())
			return
		}
	} else {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Participant added successfully",
	})
}

func RemoveChatParticipant(c *gin.Context) {
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

	participantID := c.Param("userId")
	if participantID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Participant user ID is required")
		return
	}

	client := GetCommunityServiceClient()

	// Check if user is participant in the chat
	isParticipant, err := client.IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		SendErrorResponse(c, 403, "forbidden", "You don't have access to this chat")
		return
	}

	// Use the gRPC client directly as our service client doesn't expose this method
	if CommunityClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = CommunityClient.RemoveChatParticipant(ctx, &communityProto.RemoveChatParticipantRequest{
			ChatId: chatID,
			UserId: participantID,
		})
		if err != nil {
			SendErrorResponse(c, 500, "server_error", "Failed to remove participant: "+err.Error())
			return
		}
	} else {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Participant removed successfully",
	})
}

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

func DeleteMessage(c *gin.Context) {
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

	messageID := c.Param("messageId")
	if messageID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Message ID is required")
		return
	}

	client := GetCommunityServiceClient()

	// Check if user is participant in the chat
	isParticipant, err := client.IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		SendErrorResponse(c, 403, "forbidden", "You don't have access to this chat")
		return
	}

	err = client.DeleteMessage(chatID, userID.(string), messageID)
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to delete message: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Message deleted successfully",
	})
}

func UnsendMessage(c *gin.Context) {
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

	messageID := c.Param("messageId")
	if messageID == "" {
		SendErrorResponse(c, 400, "invalid_request", "Message ID is required")
		return
	}

	client := GetCommunityServiceClient()

	// Check if user is participant in the chat
	isParticipant, err := client.IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		SendErrorResponse(c, 403, "forbidden", "You don't have access to this chat")
		return
	}

	// In a real implementation, you would fetch the specific message by ID
	// to verify the sender, but for now we'll just check chat access

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use the gRPC client directly because our service client doesn't expose UnsendMessage
	if CommunityClient != nil {
		_, err = CommunityClient.UnsendMessage(ctx, &communityProto.UnsendMessageRequest{
			MessageId: messageID,
		})
		if err != nil {
			SendErrorResponse(c, 500, "server_error", "Failed to unsend message: "+err.Error())
			return
		}
	} else {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Message unsent successfully",
	})
}

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
			// Try to get the actual user data from the user service
			var username, displayName string

			// Try using the global UserClient if available
			if UserClient != nil {
				log.Printf("Using UserClient to get info for %s", msg.SenderID)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Try getting the user data from the user service
				resp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: msg.SenderID,
				})

				if err == nil && resp != nil && resp.User != nil {
					log.Printf("Retrieved user data using UserClient for %s: username=%s name=%s",
						msg.SenderID, resp.User.Username, resp.User.Name)

					username = resp.User.Username
					displayName = resp.User.Name

					// Add profile picture if available
					if resp.User.ProfilePictureUrl != "" {
						msgObj["profile_picture_url"] = resp.User.ProfilePictureUrl
					}
				} else {
					log.Printf("Failed to get user data for %s: %v", msg.SenderID, err)
				}
			}

			// If we couldn't get real user data, generate a placeholder
			if username == "" {
				// Generate a simple name based on the sender ID
				shortenedId := msg.SenderID
				if len(shortenedId) > 4 {
					shortenedId = shortenedId[:4]
				}

				username = fmt.Sprintf("user%s", shortenedId)
				displayName = fmt.Sprintf("User %s", shortenedId)

				// Make it clearer this is generated data
				if msg.SenderID == userID.(string) {
					username = fmt.Sprintf("you_%s", username)
					displayName = fmt.Sprintf("You (%s)", displayName)
				} else {
					username = fmt.Sprintf("chat_%s", username)
					displayName = fmt.Sprintf("Chat User %s", shortenedId)
				}

				log.Printf("Using generated user data for %s: username=%s display_name=%s",
					msg.SenderID, username, displayName)
			}

			// Add the user data to the message
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

func SearchMessages(c *gin.Context) {
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

	// Parse query parameters
	query := c.Query("q")
	if query == "" {
		SendErrorResponse(c, 400, "invalid_request", "Search query is required")
		return
	}

	limit := 50 // Default limit
	offset := 0 // Default offset

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

	// Use the gRPC client directly as the service client doesn't expose search
	var messages []Message
	if CommunityClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Assuming there's a SearchMessages RPC in the proto
		resp, err := CommunityClient.SearchMessages(ctx, &communityProto.SearchMessagesRequest{
			ChatId: chatID,
			Query:  query,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			SendErrorResponse(c, 500, "server_error", "Failed to search messages: "+err.Error())
			return
		}

		// Convert proto messages to our Message type
		messages = make([]Message, len(resp.Messages))
		for i, msg := range resp.Messages {
			sentTime := time.Now()
			if msg.SentAt != nil {
				sentTime = msg.SentAt.AsTime()
			}

			messages[i] = Message{
				ID:        msg.Id,
				ChatID:    msg.ChatId,
				SenderID:  msg.SenderId,
				Content:   msg.Content,
				Timestamp: sentTime,
				IsRead:    !msg.Unsent,
				IsEdited:  false,
				IsDeleted: msg.DeletedForAll || msg.DeletedForSender,
			}
		}
	} else {
		SendErrorResponse(c, 500, "server_error", "Community service unavailable")
		return
	}

	// Format messages (similar to ListMessages)
	formattedMessages := make([]gin.H, 0, len(messages))
	for _, msg := range messages {
		// Format timestamp as Unix timestamp (seconds since epoch)
		timestamp := msg.Timestamp.Unix()

		msgObj := gin.H{
			"id":         msg.ID,
			"message_id": msg.ID,
			"chat_id":    msg.ChatID,
			"sender_id":  msg.SenderID,
			"content":    msg.Content,
			"timestamp":  timestamp,
			"user_id":    msg.SenderID,
			"is_edited":  msg.IsEdited,
			"is_deleted": msg.IsDeleted,
			"is_read":    msg.IsRead,
		}

		// Add user information (similar to ListMessages)
		if msg.SenderID != "" {
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

// GetDetailedChats returns all chats for a user with detailed participant info and last message
func GetDetailedChats(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	limit := 50 // Default limit
	offset := 0 // Default offset

	client := GetCommunityServiceClient()

	// Get the basic chat list first
	chats, err := client.GetChats(userID.(string), limit, offset)
	if err != nil {
		SendErrorResponse(c, 500, "server_error", "Failed to fetch chats: "+err.Error())
		return
	}

	// Enhanced response with detailed chats
	detailedChats := make([]gin.H, 0, len(chats))

	// Process each chat to get participants and last message
	for _, chat := range chats {
		chatID := chat.ID

		// Get participants for this chat
		participants, err := client.GetChatParticipants(chatID)
		if err != nil {
			log.Printf("Error fetching participants for chat %s: %v", chatID, err)
			continue
		}

		// Get last message for this chat
		messages, err := client.GetMessages(chatID, 1, 0)
		if err != nil {
			log.Printf("Error fetching last message for chat %s: %v", chatID, err)
		}

		// Process participant details, ideally from user service
		enhancedParticipants := make([]gin.H, len(participants))
		for i, p := range participants {
			// Try to get real user data from user service
			var username, displayName string
			var profilePicture string

			if UserClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: p,
				})

				if err == nil && resp != nil && resp.User != nil {
					username = resp.User.Username
					displayName = resp.User.Name
					profilePicture = resp.User.ProfilePictureUrl
					log.Printf("Got user data for %s: username=%s, name=%s",
						p, username, displayName)
				} else {
					log.Printf("Failed to get user data for %s: %v", p, err)
				}
			}

			// Fallback if user service fails
			if username == "" {
				shortenedId := p
				if len(shortenedId) > 4 {
					shortenedId = shortenedId[:4]
				}
				username = fmt.Sprintf("user%s", shortenedId)
				displayName = fmt.Sprintf("User %s", shortenedId)
			}

			enhancedParticipants[i] = gin.H{
				"id":                  p,
				"user_id":             p,
				"username":            username,
				"display_name":        displayName,
				"profile_picture_url": profilePicture,
			}
		}

		// Process last message if available
		var lastMessage gin.H
		if len(messages) > 0 {
			msg := messages[0]

			// Get sender info
			var senderUsername, senderDisplayName string
			var senderPicture string

			if UserClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: msg.SenderID,
				})

				if err == nil && resp != nil && resp.User != nil {
					senderUsername = resp.User.Username
					senderDisplayName = resp.User.Name
					senderPicture = resp.User.ProfilePictureUrl
				}
			}

			// Fallback if sender info not available
			if senderUsername == "" {
				shortenedId := msg.SenderID
				if len(shortenedId) > 4 {
					shortenedId = shortenedId[:4]
				}
				senderUsername = fmt.Sprintf("user%s", shortenedId)
				senderDisplayName = fmt.Sprintf("User %s", shortenedId)
			}

			lastMessage = gin.H{
				"id":         msg.ID,
				"message_id": msg.ID,
				"chat_id":    msg.ChatID,
				"sender_id":  msg.SenderID,
				"user_id":    msg.SenderID,
				"content":    msg.Content,
				"timestamp":  msg.Timestamp.Unix(),
				"is_edited":  msg.IsEdited,
				"is_deleted": msg.IsDeleted,
				"is_read":    msg.IsRead,
				"user": gin.H{
					"id":                  msg.SenderID,
					"username":            senderUsername,
					"display_name":        senderDisplayName,
					"profile_picture_url": senderPicture,
				},
			}
		}

		// Build the enhanced chat object
		detailedChat := gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"is_group_chat": chat.IsGroupChat,
			"created_by":    chat.CreatedBy,
			"created_at":    chat.CreatedAt.Unix(),
			"updated_at":    chat.UpdatedAt.Unix(),
			"participants":  enhancedParticipants,
		}

		// Add last message if available
		if lastMessage != nil {
			detailedChat["last_message"] = lastMessage
		}

		detailedChats = append(detailedChats, detailedChat)
	}

	c.JSON(200, gin.H{
		"success": true,
		"chats":   detailedChats,
	})
}

// GetChatHistoryList returns all chats for a user with the participants' usernames and the last message
func GetChatHistoryList(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, 401, "unauthorized", "Authentication required")
		return
	}

	limit := 50 // Default limit
	offset := 0 // Default offset

	client := GetCommunityServiceClient()
	log.Printf("GetChatHistoryList: Processing request for user %v", userID)

	// Get the basic chat list first
	chats, err := client.GetChats(userID.(string), limit, offset)
	if err != nil {
		log.Printf("GetChatHistoryList: Error fetching chats for user %v: %v", userID, err)
		SendErrorResponse(c, 500, "server_error", "Failed to fetch chats: "+err.Error())
		return
	}
	log.Printf("GetChatHistoryList: Found %d chats for user %v", len(chats), userID)

	// Enhanced response with detailed chat history
	chatHistoryList := make([]gin.H, 0, len(chats))

	// Process each chat to get participants and last message
	for _, chat := range chats {
		chatID := chat.ID
		log.Printf("GetChatHistoryList: Processing chat %s", chatID)

		// Get participants for this chat
		participants, err := client.GetChatParticipants(chatID)
		if err != nil {
			log.Printf("GetChatHistoryList: Error fetching participants for chat %s: %v", chatID, err)
			continue
		}
		log.Printf("GetChatHistoryList: Found %d participants for chat %s", len(participants), chatID)

		// Get last message for this chat
		messages, err := client.GetMessages(chatID, 1, 0)
		if err != nil {
			log.Printf("GetChatHistoryList: Error fetching last message for chat %s: %v", chatID, err)
		}
		log.Printf("GetChatHistoryList: Found %d messages for chat %s", len(messages), chatID)

		// Process participant details with usernames from user service
		participantDetails := make([]gin.H, len(participants))
		for i, p := range participants {
			// Get user data from user service
			var username, displayName string
			var profilePicture string

			if UserClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: p,
				})

				if err == nil && resp != nil && resp.User != nil {
					username = resp.User.Username
					displayName = resp.User.Name
					profilePicture = resp.User.ProfilePictureUrl
					log.Printf("GetChatHistoryList: Got user data for %s: username=%s", p, username)
				} else {
					log.Printf("GetChatHistoryList: Failed to get user data for %s: %v", p, err)
				}
			}

			// Fallback if user service fails
			if username == "" {
				shortenedId := p
				if len(shortenedId) > 4 {
					shortenedId = shortenedId[:4]
				}
				username = fmt.Sprintf("user%s", shortenedId)
				displayName = fmt.Sprintf("User %s", shortenedId)
				log.Printf("GetChatHistoryList: Using fallback user data for %s: username=%s", p, username)
			}

			participantDetails[i] = gin.H{
				"id":                  p,
				"user_id":             p,
				"username":            username,
				"display_name":        displayName,
				"profile_picture_url": profilePicture,
			}
		}

		// Process last message if available
		var lastMessage gin.H
		if len(messages) > 0 {
			msg := messages[0]

			// Get sender info
			var senderUsername, senderDisplayName string
			var senderPicture string

			if UserClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
					UserId: msg.SenderID,
				})

				if err == nil && resp != nil && resp.User != nil {
					senderUsername = resp.User.Username
					senderDisplayName = resp.User.Name
					senderPicture = resp.User.ProfilePictureUrl
					log.Printf("GetChatHistoryList: Got sender data for %s: username=%s", msg.SenderID, senderUsername)
				} else {
					log.Printf("GetChatHistoryList: Failed to get sender data for %s: %v", msg.SenderID, err)
				}
			}

			// Fallback if sender info not available
			if senderUsername == "" {
				shortenedId := msg.SenderID
				if len(shortenedId) > 4 {
					shortenedId = shortenedId[:4]
				}
				senderUsername = fmt.Sprintf("user%s", shortenedId)
				senderDisplayName = fmt.Sprintf("User %s", shortenedId)
				log.Printf("GetChatHistoryList: Using fallback sender data for %s: username=%s", msg.SenderID, senderUsername)
			}

			lastMessage = gin.H{
				"id":         msg.ID,
				"message_id": msg.ID,
				"chat_id":    msg.ChatID,
				"sender_id":  msg.SenderID,
				"user_id":    msg.SenderID,
				"content":    msg.Content,
				"timestamp":  msg.Timestamp.Unix(),
				"is_edited":  msg.IsEdited,
				"is_deleted": msg.IsDeleted,
				"is_read":    msg.IsRead,
				"user": gin.H{
					"id":                  msg.SenderID,
					"username":            senderUsername,
					"display_name":        senderDisplayName,
					"profile_picture_url": senderPicture,
				},
			}
		}

		// Build the chat history object
		chatHistory := gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"is_group_chat": chat.IsGroupChat,
			"created_by":    chat.CreatedBy,
			"created_at":    chat.CreatedAt.Unix(),
			"updated_at":    chat.UpdatedAt.Unix(),
			"participants":  participantDetails,
		}

		// Add last message if available
		if lastMessage != nil {
			chatHistory["last_message"] = lastMessage
		}

		chatHistoryList = append(chatHistoryList, chatHistory)
	}

	// Log the response for debugging purposes
	log.Printf("GetChatHistoryList: Returning %d chats for user %v", len(chatHistoryList), userID)

	c.JSON(200, gin.H{
		"success": true,
		"chats":   chatHistoryList,
	})
}
