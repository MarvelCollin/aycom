package handlers

import (
	"aycom/backend/api-gateway/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

func SearchUsers(c *gin.Context) {
	// Log all query parameters for debugging
	log.Printf("SearchUsers called with parameters: %v", c.Request.URL.Query())

	query := c.Query("query")
	if query == "" {
		// Try the "q" parameter as a fallback for backward compatibility
		query = c.Query("q")
		log.Printf("SearchUsers: query parameter empty, using 'q' parameter: %s", query)
	}

	// No need to check if query is empty - we want to support empty queries
	// for filter-only searches like verified users or following

	// Validate query length only if provided
	if query != "" {
		const MAX_QUERY_LENGTH = 50
		if len(query) > MAX_QUERY_LENGTH {
			log.Printf("Search query too long (%d chars), truncating to %d characters", len(query), MAX_QUERY_LENGTH)
			query = query[:MAX_QUERY_LENGTH]
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.DefaultQuery("filter", "all")

	log.Printf("SearchUsers: Processing with query='%s', filter='%s', page=%d, limit=%d", query, filter, page, limit)

	if userServiceClient == nil {
		log.Printf("SearchUsers: Error - userServiceClient is nil")
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Ensure we always have either a query or filter parameter
	// If query is empty, make sure we at least have a non-empty filter
	if query == "" && (filter == "" || filter == "all") {
		// Default to a space character for empty queries with generic filters
		query = " "
		log.Printf("SearchUsers: Empty query with generic filter, using space as query placeholder")
	}

	// Ensure we always send a filter parameter, even if it's "all"
	users, totalCount, err := userServiceClient.SearchUsers(query, filter, page, limit)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Printf("SearchUsers: gRPC error: %v, code: %v", err, st.Code())
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search users: "+st.Message())
		} else {
			log.Printf("SearchUsers: Non-gRPC error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while searching users")
		}
		log.Printf("Error searching users: %v", err)
		return
	}

	log.Printf("SearchUsers: Success, found %d users (total count: %d)", len(users), totalCount)

	var userResults []gin.H
	for _, user := range users {
		userResult := gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"is_following":        user.IsFollowing,
		}
		userResults = append(userResults, userResult)
	}

	// Calculate total pages
	totalPages := (totalCount + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	// Prepare response data
	responseData := gin.H{
		"users": userResults,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"total_pages":  totalPages,
			"has_more":     len(users) == limit && (page*limit) < totalCount,
		},
	}

	// Log the response data structure for debugging
	log.Printf("SearchUsers: Sending response data: %+v", responseData)

	utils.SendSuccessResponse(c, http.StatusOK, responseData)
}

func SearchThreads(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	// Validate query length
	const MAX_QUERY_LENGTH = 100
	if len(query) > MAX_QUERY_LENGTH {
		log.Printf("Search query too long (%d chars), truncating to %d characters", len(query), MAX_QUERY_LENGTH)
		query = query[:MAX_QUERY_LENGTH]
	}

	// Extract parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.DefaultQuery("filter", "all")
	category := c.DefaultQuery("category", "")
	sortBy := c.DefaultQuery("sort_by", "recent")
	mediaOnly := c.Query("media_only") == "true" // Check if we should only return threads with media

	// Get authenticated user ID if available
	var userID string
	userIDValue, exists := c.Get("userId")
	if exists {
		if userIDStr, ok := userIDValue.(string); ok {
			userID = userIDStr
			log.Printf("Authenticated user for search: %s", userID)
		}
	}

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	var threads []*Thread
	var err error

	log.Printf("Searching threads with query=%s, filter=%s, category=%s, sortBy=%s, mediaOnly=%v",
		query, filter, category, sortBy, mediaOnly)

	// Get threads using fuzzy search via our improved SearchThreads method
	// The method now handles the Damerau-Levenshtein fuzzy matching internally
	threads, err = threadServiceClient.SearchThreads(query, userID, page, limit)

	// Apply additional filters on the application layer
	if err == nil {
		var filteredThreads []*Thread

		// Apply following filter if needed
		if filter == "following" && userID != "" && userServiceClient != nil {
			following, err := userServiceClient.GetFollowing(userID, 1, 1000)
			if err == nil {
				// Create a map of followed user IDs for fast lookups
				followingMap := make(map[string]bool)
				for _, user := range following {
					followingMap[user.ID] = true
				}

				// Filter threads by users the current user follows
				for _, thread := range threads {
					if followingMap[thread.UserID] {
						filteredThreads = append(filteredThreads, thread)
					}
				}
				threads = filteredThreads
			} else {
				log.Printf("Error getting following users: %v", err)
			}
		}

		// Filter threads with media if mediaOnly parameter is true
		if mediaOnly {
			filteredThreads = []*Thread{}
			for _, thread := range threads {
				if len(thread.Media) > 0 {
					filteredThreads = append(filteredThreads, thread)
				}
			}
			threads = filteredThreads
		}
	}

	if err != nil {
		log.Printf("Error searching threads: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search threads")
		return
	}

	// Convert to response format
	var threadResults []gin.H
	for _, thread := range threads {
		if thread == nil {
			continue
		}

		threadData := gin.H{
			"id":                  thread.ID,
			"content":             thread.Content,
			"created_at":          thread.CreatedAt,
			"updated_at":          thread.UpdatedAt,
			"likes_count":         thread.LikeCount,
			"replies_count":       thread.ReplyCount,
			"reposts_count":       thread.RepostCount,
			"bookmark_count":      thread.BookmarkCount,
			"is_liked":            thread.IsLiked,
			"is_reposted":         thread.IsReposted,
			"is_bookmarked":       thread.IsBookmarked,
			"is_pinned":           thread.IsPinned,
			"user_id":             thread.UserID,
			"username":            thread.Username,
			"name":                thread.DisplayName,
			"profile_picture_url": thread.ProfilePicture,
		}

		// Add media if available
		if len(thread.Media) > 0 {
			var mediaList []gin.H
			for _, media := range thread.Media {
				mediaList = append(mediaList, gin.H{
					"id":   media.ID,
					"url":  media.URL,
					"type": media.Type,
				})
			}
			threadData["media"] = mediaList
		}

		threadResults = append(threadResults, threadData)
	}

	// Get the total count from the thread service (would need to be implemented)
	// For now, use a reasonable approximation
	totalCount := len(threadResults)
	totalPages := (totalCount + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	hasMore := page < totalPages

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"threads": threadResults,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"total_pages":  totalPages,
			"has_more":     hasMore,
		},
	})
}

// SearchCommunities is implemented in community_handlers.go as SearchCommunitiesHandler

func GetUserRecommendations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	users, err := userServiceClient.GetUserRecommendations(userIDStr, limit)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user recommendations: "+st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while getting user recommendations")
		}
		log.Printf("Error getting user recommendations: %v", err)
		return
	}

	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
		})
	}
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": userResults,
	})
}
