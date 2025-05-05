package routes

import (
	"aycom/backend/api-gateway/config"
	_ "aycom/backend/api-gateway/docs" // Import swagger docs
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the routes for the API Gateway
// @Summary All API Gateway Routes
// @Description Register all routes for the AYCOM platform API Gateway
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Set the config for handlers
	handlers.Config = cfg

	// Add global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// API v1 group
	v1 := router.Group("/api/v1")

	// Public routes with rate limiting
	auth := v1.Group("/auth")
	auth.Use(handlers.RateLimitMiddleware)
	{
		auth.POST("/refresh-token", handlers.RefreshToken)
		// Additional routes should use handlers that exist or be commented out
		// auth.GET("/oauth-config", handlers.GetOAuthConfig)
		// auth.POST("/login", handlers.Login)
		// auth.POST("/verify-email", handlers.VerifyEmail)
		// auth.POST("/resend-verification", handlers.ResendVerification)
		// auth.POST("/google", handlers.GoogleLogin)
	}

	// Public user registration and login
	publicUsers := v1.Group("/users")
	{
		publicUsers.POST("/register", handlers.RegisterUser)
		publicUsers.POST("/login", handlers.LoginUser)
		publicUsers.POST("/by-email", handlers.GetUserByEmail)
		// publicUsers.GET("/check-username", handlers.CheckUsernameAvailability)
		publicUsers.GET("/search", handlers.SearchUsers)
	}

	// Public thread routes
	publicThreads := v1.Group("/threads")
	publicThreads.Use(middleware.OptionalJWTAuth(string(handlers.GetJWTSecret())))
	{
		publicThreads.GET("", handlers.GetAllThreads)
		publicThreads.GET("/search", handlers.SearchThreads)
	}

	// Public trends route
	publicTrends := v1.Group("/trends")
	{
		publicTrends.GET("", handlers.GetTrends)
	}

	// Protected routes - using JWT authentication middleware
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(string(handlers.GetJWTSecret())))

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/profile", handlers.GetUserProfile)
		users.GET("/me", handlers.GetUserProfile)
		users.PUT("/profile", handlers.UpdateUserProfile)
		users.POST("/:userId/follow", handlers.FollowUser)
		users.POST("/:userId/unfollow", handlers.UnfollowUser)
		users.GET("/:userId/followers", handlers.GetFollowers)
		users.GET("/:userId/following", handlers.GetFollowing)
		users.GET("/recommendations", handlers.GetUserRecommendations)
	}

	// Thread routes
	threads := protected.Group("/threads")
	{
		threads.POST("", handlers.CreateThread)
		threads.GET("/:id", handlers.GetThread)
		threads.GET("/user/me", handlers.GetThreadsByUser)
		threads.GET("/user/:id", handlers.GetThreadsByUser)
		threads.PUT("/:id", handlers.UpdateThread)
		threads.DELETE("/:id", handlers.DeleteThread)
		// threads.POST("/media", handlers.UploadThreadMedia)

		// Social interaction routes
		threads.POST("/:id/like", handlers.LikeThread)
		threads.DELETE("/:id/like", handlers.UnlikeThread)
		threads.POST("/:id/replies", handlers.ReplyToThread)
		threads.GET("/:id/replies", handlers.GetThreadReplies)
		threads.POST("/:id/repost", handlers.RepostThread)
		threads.DELETE("/:id/repost", handlers.RemoveRepost)
		threads.POST("/:id/bookmark", handlers.BookmarkThread)
		threads.DELETE("/:id/bookmark", handlers.RemoveThreadBookmark) // Changed to match renamed function
	}

	// Product routes
	// Comment out until implemented
	/*
	products := protected.Group("/products")
	{
		products.GET("", handlers.ListProducts)
		products.GET("/:id", handlers.GetProduct)
		products.POST("", handlers.CreateProduct)
		products.PUT("/:id", handlers.UpdateProduct)
		products.DELETE("/:id", handlers.DeleteProduct)
	}
	*/

	// Community routes
	communities := protected.Group("/communities")
	{
		communities.POST("", handlers.CreateCommunity)
		communities.GET("", handlers.ListCommunities)
		communities.GET("/:id", handlers.GetCommunityByID)
		communities.PUT("/:id", handlers.UpdateCommunity)
		communities.DELETE("/:id", handlers.DeleteCommunity)
		communities.POST("/:id/approve", handlers.ApproveCommunity)
		communities.GET("/search", handlers.SearchCommunities)

		communities.POST("/:id/members", handlers.AddMember)
		communities.GET("/:id/members", handlers.ListMembers)
		communities.PUT("/:id/members/:userId", handlers.UpdateMemberRole)
		communities.DELETE("/:id/members/:userId", handlers.RemoveMember)

		communities.POST("/:id/rules", handlers.AddRule)
		communities.GET("/:id/rules", handlers.ListRules)
		communities.DELETE("/:id/rules/:ruleId", handlers.RemoveRule)

		communities.POST("/:id/join-requests", handlers.RequestToJoin)
		communities.GET("/:id/join-requests", handlers.ListJoinRequests)
		communities.POST("/:id/join-requests/:requestId/approve", handlers.ApproveJoinRequest)
		communities.POST("/:id/join-requests/:requestId/reject", handlers.RejectJoinRequest)
	}

	// Chat routes - comment out until implemented
	// Chat routes
	chats := protected.Group("/chats")
	{
		chats.POST("", handlers.CreateChat)
		chats.GET("", handlers.ListChats)
		chats.GET("/:id/participants", handlers.ListChatParticipants)
		chats.POST("/:id/participants", handlers.AddChatParticipant)
		chats.DELETE("/:id/participants/:userId", handlers.RemoveChatParticipant)

		chats.POST("/:id/messages", handlers.SendMessage)
		chats.GET("/:id/messages", handlers.ListMessages)
		chats.DELETE("/:id/messages/:messageId", handlers.DeleteMessage)
		chats.POST("/:id/messages/:messageId/unsend", handlers.UnsendMessage)
		chats.GET("/:id/messages/search", handlers.SearchMessages)

		// WebSocket endpoint for real-time chat
		chats.GET("/:id/ws", handlers.HandleCommunityChat)
	}

	// Notification routes
	notifications := protected.Group("/notifications")
	{
		notifications.GET("", handlers.GetUserNotifications)
		notifications.GET("/mentions", handlers.GetMentionNotifications)
		notifications.POST("/:id/read", handlers.MarkNotificationAsRead)
		notifications.POST("/read-all", handlers.MarkAllNotificationsAsRead)
		notifications.DELETE("/:id", handlers.DeleteNotification)
		notifications.GET("/ws", handlers.HandleNotificationsWebSocket)
	}

	// Bookmarks routes - new group
	bookmarks := protected.Group("/bookmarks")
	{
		bookmarks.GET("", handlers.GetUserBookmarks)
		bookmarks.GET("/search", handlers.SearchBookmarks)
	}

	// Media routes - new group for general media upload
	media := protected.Group("/media")
	{
		media.POST("", handlers.UploadMedia)
		media.GET("/search", handlers.SearchMedia)
	}
}
