package routes

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"aycom/backend/api-gateway/config"
	_ "aycom/backend/api-gateway/docs"
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"
)

func CORSPreflightHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3000"
		}

		log.Printf("CORS Preflight for %s: Setting Allow-Origin to %s", c.Request.URL.Path, origin)

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Powered-By")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	handlers.AppConfig = cfg

	v1 := router.Group("/api/v1")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "wompwompAWIKWOKKWOKWOK"
	}

	v1.GET("/public-suggestions", handlers.GetPublicUserSuggestions)

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
		auth.GET("/check-admin", middleware.JWTAuth(jwtSecret), handlers.CheckAdminStatus)
	}

	ai := v1.Group("/ai")
	{
		ai.POST("/predict-category", handlers.PredictCategory)
	}

	publicUsers := v1.Group("/users")

	publicUsers.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		publicUsers.POST("/register", handlers.RegisterUser)
		publicUsers.POST("/login", handlers.LoginUser)
		publicUsers.POST("/by-email", handlers.GetUserByEmail)
		publicUsers.GET("/check-username", handlers.CheckUsernameAvailability)
		publicUsers.GET("/username/:username", handlers.GetUserByUsername)
		publicUsers.GET("/:userId", handlers.GetUserById)
		publicUsers.GET("/search", handlers.SearchUsers)
		publicUsers.GET("/all", handlers.GetAllUsers)
		publicUsers.GET("", handlers.GetAllUsers)

		publicUsers.POST("/admin/create", handlers.CreateAdminUser)
	}

	v1.OPTIONS("/threads", CORSPreflightHandler())
	v1.OPTIONS("/threads/*path", CORSPreflightHandler())

	// Add thread related endpoints
	threads := v1.Group("/threads")
	threads.Use(middleware.JWTAuth(jwtSecret))
	{
		threads.POST("", handlers.CreateThread)
		threads.PUT("/:id", handlers.UpdateThread)
		threads.DELETE("/:id", handlers.DeleteThread)
		threads.POST("/:id/like", handlers.LikeThread)
		threads.DELETE("/:id/like", handlers.UnlikeThread)
		threads.POST("/:id/bookmark", handlers.BookmarkThread)
		threads.DELETE("/:id/bookmark", handlers.RemoveBookmark)
		threads.GET("/:id/replies", handlers.GetThreadReplies)
		threads.POST("/:id/replies", handlers.ReplyToThread)
		threads.POST("/:id/repost", handlers.RepostThread)
		threads.DELETE("/:id/repost", handlers.RemoveRepost)
		threads.POST("/:id/pin", handlers.PinThread)
		threads.DELETE("/:id/pin", handlers.UnpinThread)
	}

	// Public thread routes with optional authentication
	publicThreads := v1.Group("/threads")
	publicThreads.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		publicThreads.GET("", handlers.GetAllThreads)
		publicThreads.GET("/search", handlers.SearchThreads)
		publicThreads.GET("/trending", handlers.GetTrends)
		publicThreads.GET("/following", handlers.GetThreadsFromFollowing)
		publicThreads.GET("/hashtag/:hashtag", handlers.GetThreadsByHashtag)
		publicThreads.GET("/:id", handlers.GetThread)
	}

	// Public search routes with optional authentication
	search := v1.Group("/search")
	search.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		search.GET("/threads", handlers.SearchThreads)
		search.GET("/users", handlers.SearchUsers)
		search.GET("/communities", handlers.OldSearchCommunities)
		search.GET("/media", handlers.SearchMedia)
	}

	trendsGroup := v1.Group("/trends")
	trendsGroup.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		trendsGroup.GET("", handlers.GetTrends)
	}

	v1.Group("/categories").GET("", handlers.GetCategories)

	v1.GET("/communities/categories", handlers.ListCategories)

	v1.GET("/communities/search", handlers.OldSearchCommunities)

	// Create a public communities group with optional JWT auth for user-specific endpoints
	publicCommunities := v1.Group("/communities")
	publicCommunities.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		publicCommunities.GET("/user/:userId/joined", handlers.GetJoinedCommunities)
		publicCommunities.GET("/user/:userId/pending", handlers.GetPendingCommunities)
		publicCommunities.GET("/discover", handlers.GetDiscoverCommunities)
		publicCommunities.GET("", handlers.ListCommunities)
		publicCommunities.GET("/:id", handlers.GetCommunityByID)
	}

	publicWebsockets := v1.Group("/chats")
	{
		publicWebsockets.GET("/:id/ws", handlers.HandleChatWebSocket)
	}

	users := v1.Group("/users")
	users.Use(middleware.JWTAuth(jwtSecret))
	{
		users.GET("/profile", handlers.GetUserProfile)
		users.GET("/me", handlers.GetUserProfile)
		users.PUT("/profile", handlers.UpdateUserProfile)
		users.POST("/media", handlers.UploadProfileMedia)
		users.POST("/profile-picture/update", handlers.UpdateProfilePictureURLHandler)
		users.POST("/banner/update", handlers.UpdateBannerURLHandler)
		users.POST("/:userId/follow", handlers.FollowUser)
		users.POST("/:userId/unfollow", handlers.UnfollowUser)
		users.GET("/:userId/followers", handlers.GetFollowers)
		users.GET("/:userId/following", handlers.GetFollowing)
		users.GET("/:userId/follow-status", handlers.CheckFollowStatus)
		users.GET("/recommendations", handlers.GetUserRecommendations)
		users.POST("/admin-status", handlers.UpdateUserAdminStatus)
		users.POST("/:userId/block", handlers.BlockUser)
		users.POST("/:userId/unblock", handlers.UnblockUser)
		users.GET("/blocked", handlers.GetBlockedUsers)
		users.POST("/:userId/report", handlers.ReportUser)
		users.POST("/premium-request", handlers.CreatePremiumRequest)
	}

	replies := v1.Group("/replies")
	replies.Use(middleware.JWTAuth(jwtSecret))
	{
		replies.POST("/:id/like", handlers.LikeReply)
		replies.DELETE("/:id/like", handlers.UnlikeReply)
		replies.POST("/:id/bookmark", handlers.BookmarkReply)
		replies.DELETE("/:id/bookmark", handlers.RemoveReplyBookmark)
		replies.POST("/:id/replies", handlers.ReplyToThread)
		replies.POST("/:id/pin", handlers.PinReply)
		replies.DELETE("/:id/pin", handlers.UnpinReply)
	}

	// Add a public version of replies endpoint with optional authentication
	publicReplies := v1.Group("/replies")
	publicReplies.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		publicReplies.GET("/:id/replies", handlers.GetRepliesByParentReply)
	}

	communities := v1.Group("/communities")
	communities.Use(middleware.JWTAuth(jwtSecret))
	{
		communities.POST("", handlers.CreateCommunity)
		communities.PUT("/:id", handlers.UpdateCommunity)
		communities.DELETE("/:id", handlers.DeleteCommunity)
		communities.POST("/:id/approve", handlers.ApproveCommunity)

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

		communities.GET("/:id/membership", handlers.CheckMembershipStatus)
	}

	communityDetails := v1.Group("/communities/:id")
	communityDetails.Use(middleware.OptionalJWTAuth(jwtSecret))
	{
		communityDetails.GET("/top-members", handlers.GetTopCommunityMembers)
		communityDetails.GET("/threads/top", handlers.GetCommunityThreadsByLikes)
		communityDetails.GET("/threads/latest", handlers.GetCommunityThreadsByDate)
		communityDetails.GET("/threads/media", handlers.GetCommunityMediaThreads)
	}

	chats := v1.Group("/chats")
	chats.Use(middleware.JWTAuth(jwtSecret))
	{
		chats.POST("", handlers.CreateChat)
		chats.GET("", handlers.ListChats)
		chats.GET("/history", handlers.GetChatHistoryList)
		chats.GET("/:id/participants", handlers.ListChatParticipants)
		chats.POST("/:id/participants", handlers.AddChatParticipant)
		chats.DELETE("/:id/participants/:userId", handlers.RemoveChatParticipant)
		chats.DELETE("/:id", handlers.DeleteChat)

		chats.POST("/:id/messages", handlers.SendMessage)
		chats.GET("/:id/messages", handlers.ListMessages)
		chats.DELETE("/:id/messages/:messageId", handlers.DeleteMessage)
		chats.POST("/:id/messages/:messageId/unsend", handlers.UnsendMessage)
		chats.GET("/:id/messages/search", handlers.SearchMessages)
	}

	notifications := v1.Group("/notifications")
	notifications.Use(middleware.JWTAuth(jwtSecret))
	{
		notifications.GET("", handlers.GetUserNotifications)
		notifications.GET("/interactions", handlers.GetUserInteractionNotifications)
		notifications.GET("/mentions", handlers.GetMentionNotifications)
		notifications.POST("/:id/read", handlers.MarkNotificationAsRead)
		notifications.POST("/read-all", handlers.MarkAllNotificationsAsRead)
		notifications.DELETE("/:id", handlers.DeleteNotification)
	}

	// WebSocket endpoint without JWT middleware (handles auth internally)
	v1.GET("/notifications/ws", handlers.HandleNotificationsWebSocket)

	bookmarks := v1.Group("/bookmarks")
	bookmarks.Use(middleware.JWTAuth(jwtSecret))
	{
		bookmarks.GET("", handlers.GetUserBookmarks)
	}
	media := v1.Group("/media")
	media.Use(middleware.JWTAuth(jwtSecret))
	{
		media.POST("", handlers.UploadMedia)
		media.GET("/search", handlers.SearchMedia)
	}

	v1.OPTIONS("/admin/*path", func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3000"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token, Pragma, Expires, Connection, User-Agent, Host, Referer, Cookie, Set-Cookie, *")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Powered-By")

		c.AbortWithStatus(http.StatusNoContent)
	})

	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(jwtSecret))
	adminGroup.Use(middleware.AdminOnly())
	{
		adminGroup.GET("/dashboard/statistics", handlers.GetDashboardStatistics)
		adminGroup.POST("/users/:userId/ban", handlers.BanUser)
		adminGroup.POST("/newsletter/send", handlers.SendNewsletter)
		adminGroup.GET("/community-requests", handlers.GetCommunityRequests)
		adminGroup.POST("/community-requests/:requestId/process", handlers.ProcessCommunityRequest)
		adminGroup.POST("/community-requests/sync", handlers.SyncPendingCommunities)
		adminGroup.GET("/premium-requests", handlers.GetPremiumRequests)
		adminGroup.POST("/premium-requests/:requestId/process", handlers.ProcessPremiumRequest)
		adminGroup.GET("/report-requests", handlers.GetReportRequests)
		adminGroup.POST("/report-requests/:requestId/process", handlers.ProcessReportRequest)
		adminGroup.GET("/thread-categories", handlers.GetThreadCategories)
		adminGroup.POST("/thread-categories", handlers.CreateThreadCategory)
		adminGroup.PUT("/thread-categories/:categoryId", handlers.UpdateThreadCategory)
		adminGroup.DELETE("/thread-categories/:categoryId", handlers.DeleteThreadCategory)
		adminGroup.GET("/community-categories", handlers.GetCommunityCategories)
		adminGroup.POST("/community-categories", handlers.CreateCommunityCategory)
		adminGroup.PUT("/community-categories/:categoryId", handlers.UpdateCommunityCategory)
		adminGroup.DELETE("/community-categories/:categoryId", handlers.DeleteCommunityCategory)
		adminGroup.GET("/newsletter-subscribers", handlers.AdminGetAllUsers)
	}
}
