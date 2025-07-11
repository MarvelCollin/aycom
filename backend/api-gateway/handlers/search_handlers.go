package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

func SearchUsers(c *gin.Context) {
	
	query := c.Query("q")
	filter := c.DefaultQuery("filter", "all")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	fuzzyStr := c.DefaultQuery("fuzzy", "false")

	
	log.Printf("SearchUsers: Processing with query='%s', filter='%s', page=%s, limit=%s, fuzzy=%s",
		query, filter, pageStr, limitStr, fuzzyStr)

	
	enableFuzzy := fuzzyStr == "true"

	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	
	if query == "" {
		query = c.Query("query")
		log.Printf("SearchUsers: query parameter empty, using 'q' parameter: %s", query)
	}

	
	

	
	if query != "" {
		const MAX_QUERY_LENGTH = 50
		if len(query) > MAX_QUERY_LENGTH {
			log.Printf("Search query too long (%d chars), truncating to %d characters", len(query), MAX_QUERY_LENGTH)
			query = query[:MAX_QUERY_LENGTH]
		}
	}

	if userServiceClient == nil {
		log.Printf("SearchUsers: Error - userServiceClient is nil")
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	
	
	if query == "" && (filter == "" || filter == "all") {
		
		query = " "
		log.Printf("SearchUsers: Empty query with generic filter, using space as query placeholder")
	}

	
	users, totalCount, searchErr := userServiceClient.SearchUsers(query, filter, page, limit, enableFuzzy)

	if searchErr != nil {
		st, ok := status.FromError(searchErr)
		if ok {
			log.Printf("SearchUsers: gRPC error: %v, code: %v", searchErr, st.Code())
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search users: "+st.Message())
		} else {
			log.Printf("SearchUsers: Non-gRPC error: %v", searchErr)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while searching users")
		}
		log.Printf("Error searching users: %v", searchErr)
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

	
	totalPages := (totalCount + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	
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

	
	log.Printf("SearchUsers: Sending response data: %+v", responseData)

	utils.SendSuccessResponse(c, http.StatusOK, responseData)
}

func SearchThreads(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	
	const MAX_QUERY_LENGTH = 100
	if len(query) > MAX_QUERY_LENGTH {
		log.Printf("Search query too long (%d chars), truncating to %d characters", len(query), MAX_QUERY_LENGTH)
		query = query[:MAX_QUERY_LENGTH]
	}

	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.DefaultQuery("filter", "all")
	category := c.DefaultQuery("category", "")
	sortBy := c.DefaultQuery("sort_by", "recent")
	mediaOnly := c.Query("media_only") == "true" 

	
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

	
	
	threads, err = threadServiceClient.SearchThreads(query, userID, page, limit)

	
	if err == nil {
		var filteredThreads []*Thread

		
		if filter == "following" && userID != "" && userServiceClient != nil {
			following, err := userServiceClient.GetFollowing(userID, 1, 1000)
			if err == nil {
				
				followingMap := make(map[string]bool)
				for _, user := range following {
					followingMap[user.ID] = true
				}

				
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
