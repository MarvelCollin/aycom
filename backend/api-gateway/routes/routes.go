package routes

import (
	"aycom/backend/api-gateway/config"
	_ "aycom/backend/api-gateway/docs"
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	handlers.AppConfig = cfg

	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	v1 := router.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Use(handlers.RateLimitMiddleware)
	{
		auth.POST("/refresh-token", handlers.RefreshToken)
		auth.GET("/oauth-config", handlers.GetOAuthConfig)
		auth.POST("/login", handlers.Login)
		auth.POST("/verify-email", handlers.VerifyEmail)
		auth.POST("/resend-verification", handlers.ResendVerification)
		auth.POST("/google", handlers.GoogleLogin)
		auth.POST("/forgot-password", handlers.ForgotPassword)
		auth.POST("/verify-security-answer", handlers.VerifySecurityAnswer)
		auth.POST("/reset-password", handlers.ResetPassword)
	}

	ai := v1.Group("/ai")
	{
		ai.POST("/predict-category", handlers.PredictCategory)
	}

	publicUsers := v1.Group("/users")
	{
		publicUsers.POST("/register", handlers.RegisterUser)
		publicUsers.POST("/login", handlers.LoginUser)
		publicUsers.POST("/by-email", handlers.GetUserByEmail)
		publicUsers.GET("/check-username", handlers.CheckUsernameAvailability)
		publicUsers.GET("/search", handlers.SearchUsers)
		publicUsers.GET("/suggestions", handlers.GetUserSuggestions)
		publicUsers.GET("/all", handlers.GetAllUsers)
	}

	publicThreads := v1.Group("/threads")
	publicThreads.Use(middleware.OptionalJWTAuth(string(handlers.GetJWTSecret())))
	{
		publicThreads.GET("", handlers.GetAllThreads)
		publicThreads.GET("/search", handlers.SearchThreads)
	}

	v1.Group("/trends").GET("", handlers.GetTrends)

	v1.Group("/categories").GET("", handlers.GetCategories)

	publicWebsockets := v1.Group("/chats")
	{
		publicWebsockets.GET("/:id/ws", handlers.HandleCommunityChat)
	}

	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(string(handlers.GetJWTSecret())))

	users := protected.Group("/users")
	{
		users.GET("/profile", handlers.GetUserProfile)
		users.GET("/me", handlers.GetUserProfile)
		users.PUT("/profile", handlers.UpdateUserProfile)
		users.POST("/media", handlers.UploadProfileMedia)
		users.POST("/:userId/follow", handlers.FollowUser)
		users.POST("/:userId/unfollow", handlers.UnfollowUser)
		users.GET("/:userId/followers", handlers.GetFollowers)
		users.GET("/:userId/following", handlers.GetFollowing)
		users.GET("/recommendations", handlers.GetUserRecommendations)
	}

	threads := protected.Group("/threads")
	{
		threads.POST("", handlers.CreateThread)
		threads.GET("/:id", handlers.GetThread)
		threads.GET("/user/me", handlers.GetThreadsByUser)
		threads.GET("/user/:id", handlers.GetThreadsByUser)
		threads.GET("/user/:id/replies", handlers.GetUserReplies)
		threads.GET("/user/:id/likes", handlers.GetUserLikedThreads)
		threads.GET("/user/:id/media", handlers.GetUserMedia)
		threads.PUT("/:id", handlers.UpdateThread)
		threads.DELETE("/:id", handlers.DeleteThread)
		threads.GET("/following", handlers.GetThreadsFromFollowing)
		threads.POST("/media", handlers.UploadThreadMedia)
		threads.POST("/:id/pin", handlers.PinThread)
		threads.DELETE("/:id/pin", handlers.UnpinThread)

		threads.POST("/:id/like", handlers.LikeThread)
		threads.DELETE("/:id/like", handlers.UnlikeThread)
		threads.POST("/:id/replies", handlers.ReplyToThread)
		threads.GET("/:id/replies", handlers.GetThreadReplies)
		threads.POST("/:id/repost", handlers.RepostThread)
		threads.DELETE("/:id/repost", handlers.RemoveRepost)
		threads.POST("/:id/bookmark", handlers.BookmarkThreadHandler)
		threads.DELETE("/:id/bookmark", handlers.RemoveBookmark)
	}

	replies := protected.Group("/replies")
	{
		replies.POST("/:id/like", handlers.LikeReply)
		replies.DELETE("/:id/like", handlers.UnlikeReply)
		replies.POST("/:id/bookmark", handlers.BookmarkReply)
		replies.DELETE("/:id/bookmark", handlers.RemoveReplyBookmark)
		replies.POST("/:id/replies", handlers.ReplyToThread)
		replies.GET("/:id/replies", handlers.GetRepliesByParentReply)
		replies.POST("/:id/pin", handlers.PinReply)
		replies.DELETE("/:id/pin", handlers.UnpinReply)
	}

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

	chats := protected.Group("/chats")
	{
		chats.POST("", handlers.CreateChat)
		chats.GET("", handlers.ListChats)
		chats.GET("/history", handlers.GetChatHistoryList)
		chats.GET("/:id/participants", handlers.ListChatParticipants)
		chats.POST("/:id/participants", handlers.AddChatParticipant)
		chats.DELETE("/:id/participants/:userId", handlers.RemoveChatParticipant)

		chats.POST("/:id/messages", handlers.SendMessage)
		chats.GET("/:id/messages", handlers.ListMessages)
		chats.DELETE("/:id/messages/:messageId", handlers.DeleteMessage)
		chats.POST("/:id/messages/:messageId/unsend", handlers.UnsendMessage)
		chats.GET("/:id/messages/search", handlers.SearchMessages)
	}

	notifications := protected.Group("/notifications")
	{
		notifications.GET("", handlers.GetUserNotifications)
		notifications.GET("/mentions", handlers.GetMentionNotifications)
		notifications.POST("/:id/read", handlers.MarkNotificationAsRead)
		notifications.POST("/read-all", handlers.MarkAllNotificationsAsRead)
		notifications.DELETE("/:id", handlers.DeleteNotification)
		notifications.GET("/ws", handlers.HandleNotificationsWebSocket)
	}

	bookmarks := protected.Group("/bookmarks")
	{
		bookmarks.GET("", handlers.GetUserBookmarks)
		bookmarks.GET("/search", handlers.SearchBookmarks)
		bookmarks.DELETE("/:id", handlers.DeleteBookmarkById)
	}

	media := protected.Group("/media")
	{
		media.POST("", handlers.UploadMedia)
		media.GET("/search", handlers.SearchMedia)
	}
}
