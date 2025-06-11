package handlers

import (
	"aycom/backend/api-gateway/utils"
	communityProto "aycom/backend/proto/community"
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Return a response format the frontend expects
	c.JSON(201, gin.H{
		"chat": gin.H{
			"id":            chat.ID,
			"name":          chat.Name,
			"type":          "individual",
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

	// Log the request for debugging
	log.Printf("ListMessages request: chatID=%s, userID=%v", chatID, userID)

	// Validate UUID format
	_, uuidErr := uuid.Parse(chatID)
	if uuidErr != nil {
		log.Printf("Invalid chat UUID format: %v", uuidErr)
		utils.SendErrorResponse(c, 400, "INVALID_ID", "Invalid chat ID format")
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
		log.Printf("ERROR: CommunityClient is nil in ListMessages")
		utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, verify that the user has access to this chat
	isParticipant, err := GetCommunityServiceClient().IsUserChatParticipant(chatID, userID.(string))
	if err != nil {
		log.Printf("Error checking chat participation: %v", err)
		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to verify chat access: "+err.Error())
		return
	}

	if !isParticipant {
		log.Printf("User %s is not a participant in chat %s", userID, chatID)
		utils.SendErrorResponse(c, 403, "FORBIDDEN", "You are not a participant in this chat")
		return
	}

	// List messages
	log.Printf("Fetching messages for chat %s (limit: %d, offset: %d)", chatID, limit, offset)
	resp, err := CommunityClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
		ChatId: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		log.Printf("Error listing messages: %v", err)

		// Handle specific gRPC errors with appropriate responses
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				utils.SendErrorResponse(c, 404, "NOT_FOUND", "Chat not found")
				return
			case codes.InvalidArgument:
				utils.SendErrorResponse(c, 400, "INVALID_REQUEST", st.Message())
				return
			case codes.Unavailable:
				utils.SendErrorResponse(c, 503, "SERVICE_UNAVAILABLE", "Community service is unavailable")
				return
			default:
				utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list messages: "+err.Error())
				return
			}
		}

		utils.SendErrorResponse(c, 500, "SERVER_ERROR", "Failed to list messages: "+err.Error())
		return
	}

	// If we got no response or no messages, return an empty array
	if resp == nil {
		log.Printf("Warning: ListMessages returned nil response for chat %s", chatID)
		utils.SendSuccessResponse(c, 200, gin.H{
			"messages": []gin.H{},
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  0,
			},
		})
		return
	}

	// Initialize messages to empty array if nil
	messages := []gin.H{}
	if resp.Messages != nil {
		messages = make([]gin.H, 0, len(resp.Messages))

		for _, msg := range resp.Messages {
			// Skip nil messages
			if msg == nil {
				continue
			}

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
	}

	log.Printf("Successfully retrieved %d messages for chat %s", len(messages), chatID)
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
