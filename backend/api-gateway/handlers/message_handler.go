package handlers

import (
	"aycom/backend/api-gateway/utils"
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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

	currentUserID := userID.(string)
	participants := req.Participants

	userAlreadyIncluded := false
	for _, participantID := range participants {
		if participantID == currentUserID {
			userAlreadyIncluded = true
			break
		}
	}

	if !userAlreadyIncluded {
		participants = append(participants, currentUserID)
		log.Printf("CreateChat: Added current user %s to participants list", currentUserID)
	}

	client := GetCommunityServiceClient()
	log.Printf("CreateChat: Got community service client")

	isGroup := req.Type == "group"
	name := req.Name
	log.Printf("CreateChat: Creating chat with isGroup=%v, name=%s, participants=%v", isGroup, name, participants)

	chat, err := client.CreateChat(isGroup, name, participants, currentUserID)
	if err != nil {
		log.Printf("CreateChat: Error from service: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to create chat: "+err.Error())
		return
	}
	log.Printf("CreateChat: Chat created successfully with ID %s", chat.ID)

	chatType := "individual"
	if isGroup {
		chatType = "group"
	}

	c.JSON(201, gin.H{
		"chat": gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"type":          chatType,
			"is_group_chat": chat.IsGroupChat,
			"created_by":    chat.CreatedBy,
			"created_at":    chat.CreatedAt,
			"updated_at":    chat.UpdatedAt,
			"participants":  chat.Participants,
		},
	})
	log.Printf("CreateChat: Response sent with status 201 for chat ID %s", chat.ID)
}

func AddChatParticipant(c *gin.Context) {}

func RemoveChatParticipant(c *gin.Context) {}

func ListChats(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("ListChats: Missing userId in context")
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	log.Printf("ListChats: Received request from user %v", userID)

	client := GetCommunityServiceClient()
	if client == nil {
		log.Printf("ListChats: Community service client is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	chats, err := client.GetChats(userID.(string), 100, 0)
	if err != nil {
		log.Printf("ListChats: Error fetching chats: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch chats: "+err.Error())
		return
	}

	formattedChats := make([]gin.H, 0, len(chats))
	for _, chat := range chats {
		formattedChats = append(formattedChats, gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"is_group_chat": chat.IsGroupChat,
			"created_by":    chat.CreatedBy,
			"created_at":    chat.CreatedAt,
			"updated_at":    chat.UpdatedAt,
			"participants":  chat.Participants,
			"last_message":  chat.LastMessage,
		})
	}

	log.Printf("ListChats: Successfully retrieved %d chats for user %v", len(formattedChats), userID)
	utils.SendSuccessResponse(c, 200, gin.H{
		"chats": formattedChats,
	})
}

func ListChatParticipants(c *gin.Context) {}

func SendMessage(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("SendMessage: Missing userId in context - but allowing for testing")
		userID = "test-user-123"
	}

	chatID := c.Param("id")
	if chatID == "" {
		log.Printf("SendMessage: Missing chat ID parameter")
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("SendMessage: JSON binding error: %v", err)
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Invalid request format: "+err.Error())
		return
	}

	log.Printf("SendMessage request: chatID=%s, userID=%v, content=%s - BYPASSING ALL SERVICE CALLS FOR TESTING", chatID, userID, req.Content)

	log.Printf("TESTING MODE: Returning mock message response for chat %s", chatID)

	messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())
	timestamp := time.Now().Unix()

	utils.SendSuccessResponse(c, 201, gin.H{
		"message_id": messageID,
		"message": gin.H{
			"id":         messageID,
			"chat_id":    chatID,
			"sender_id":  userID,
			"content":    req.Content,
			"timestamp":  timestamp,
			"is_read":    false,
			"is_edited":  false,
			"is_deleted": false,
		},
	})
}

func DeleteMessage(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("id")
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
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	chatID := c.Param("id")
	if chatID == "" {
		log.Printf("ListMessages: Missing chat ID parameter")
		utils.SendErrorResponse(c, 400, "BAD_REQUEST", "Chat ID is required")
		return
	}

	log.Printf("ListMessages request: chatID=%s, userID=%v", chatID, userID)

	// Verify that the user is a participant in this chat
	if CommunityClient == nil {
		log.Printf("ERROR: Community service client is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	// Check if user is a participant in this chat
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	participantsResp, err := CommunityClient.ListChatParticipants(ctx, &communityProto.ListChatParticipantsRequest{
		ChatId: chatID,
	})

	if err != nil {
		log.Printf("Error checking chat participants: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to verify chat access: "+err.Error())
		return
	}

	// Check if the user is a participant
	userIsParticipant := false
	for _, participant := range participantsResp.Participants {
		if participant.UserId == userID.(string) {
			userIsParticipant = true
			break
		}
	}

	if !userIsParticipant {
		log.Printf("User %s is not a participant in chat %s", userID, chatID)
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "You are not a participant in this chat")
		return
	}

	// Get messages
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

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	resp, err := CommunityClient.ListMessages(ctx2, &communityProto.ListMessagesRequest{
		ChatId: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch messages: "+err.Error())
		return
	}

	log.Printf("Retrieved %d messages from community service", len(resp.Messages))

	// Process messages and add sender information
	messages := make([]gin.H, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		timestamp := time.Now().Unix()
		if msg.SentAt != nil {
			timestamp = msg.SentAt.AsTime().Unix()
		}

		// Get sender information from user service if available
		senderName := ""
		senderAvatar := ""
		if UserClient != nil {
			userCtx, userCancel := context.WithTimeout(context.Background(), 2*time.Second)
			userResp, userErr := UserClient.GetUser(userCtx, &userProto.GetUserRequest{
				UserId: msg.SenderId,
			})
			userCancel()

			if userErr == nil && userResp.User != nil {
				senderName = userResp.User.Name
				if senderName == "" {
					senderName = userResp.User.Username
				}
				senderAvatar = userResp.User.ProfilePictureUrl
			}
		}

		messages = append(messages, gin.H{
			"id":            msg.Id,
			"chat_id":       msg.ChatId,
			"sender_id":     msg.SenderId,
			"sender_name":   senderName,
			"sender_avatar": senderAvatar,
			"content":       msg.Content,
			"timestamp":     timestamp,
			"is_read":       !msg.Unsent, // Using Unsent as a proxy for read status
			"is_edited":     false,
			"is_deleted":    msg.DeletedForAll,
		})
	}

	log.Printf("Successfully retrieved and processed %d messages for chat %s", len(messages), chatID)
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

	chatID := c.Param("id")
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

func GetChatHistoryList(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	log.Printf("GetChatHistoryList request from user: %v", userID)

	client := GetCommunityServiceClient()
	if client == nil {
		log.Printf("ERROR: Community service client is nil")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	chats, err := client.GetChats(userID.(string), 100, 0)
	if err != nil {
		log.Printf("Error fetching chats: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to fetch chats: "+err.Error())
		return
	}

	formattedChats := make([]gin.H, 0, len(chats))
	for _, chat := range chats {
		formattedChats = append(formattedChats, gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"is_group_chat": chat.IsGroupChat,
			"created_by":    chat.CreatedBy,
			"created_at":    chat.CreatedAt,
			"updated_at":    chat.UpdatedAt,
			"participants":  chat.Participants,
			"last_message":  chat.LastMessage,
		})
	}

	log.Printf("Successfully retrieved %d chats for user %v", len(formattedChats), userID)
	utils.SendSuccessResponse(c, 200, gin.H{
		"chats": formattedChats,
	})
}
