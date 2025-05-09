package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	threadProto "aycom/backend/proto/thread"

	"github.com/gin-gonic/gin"
)

type Trend struct {
	ID        string `json:"id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	PostCount int    `json:"post_count"`
}

// Cache trends for performance with 10-minute expiration
var (
	trendsCache     []Trend
	trendsCacheLock sync.RWMutex
	lastUpdated     time.Time
	cacheExpiration = 10 * time.Minute
)

// fetchTrends retrieves trending topics from the thread service
func fetchTrends() []Trend {
	log.Println("Fetching trending hashtags from thread service")

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("Failed to connect to thread service: %v", err)
		return []Trend{} // Return empty trends on error
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call thread service to get trending hashtags
	resp, err := client.GetTrendingHashtags(ctx, &threadProto.GetTrendingHashtagsRequest{
		Limit: 20, // Request more than we need to have a good selection
	})
	if err != nil {
		log.Printf("Failed to get trending hashtags: %v", err)
		return []Trend{} // Return empty trends on error
	}

	// Convert hashtag responses to trends
	trends := make([]Trend, 0, len(resp.Hashtags))
	for _, hashtag := range resp.Hashtags {
		trends = append(trends, Trend{
			ID:        hashtag.Name,
			Category:  "Hashtag",
			Title:     "#" + hashtag.Name,
			PostCount: int(hashtag.Count),
		})
	}

	return trends
}

// @Summary Get trends and user recommendations for the home feed
// @Description Returns trending hashtags and optionally user recommendations for the home feed
// @Tags Trends,Home Feed
// @Produce json
// @Param limit query int false "Number of trends to return (default 5)"
// @Param include_recommendations query bool false "Whether to include user recommendations in the response"
// @Param rec_limit query int false "Number of user recommendations to return (default 3, requires include_recommendations=true)"
// @Success 200 {object} models.TrendsResponse
// @Router /api/v1/trends [get]
func GetTrends(c *gin.Context) {
	// Check if the route was accessed through the authenticated route
	// If yes, validate the authentication, otherwise continue
	var userIDStr string
	if userID, exists := c.Get("userId"); !exists {
		// If the userId doesn't exist in context, we're likely in a public endpoint
		// Just log the info and continue with the request
		log.Println("Trends accessed via public route")
	} else {
		log.Printf("Trends accessed by authenticated user: %v", userID)
		userIDStr = userID.(string)
	}

	// Get trending hashtags limit, default to 5
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

	// Use cache if not expired
	trendsCacheLock.RLock()
	cacheValid := !lastUpdated.IsZero() && time.Since(lastUpdated) < cacheExpiration
	trendsCacheLock.RUnlock()

	var trends []Trend
	if cacheValid {
		trendsCacheLock.RLock()
		trends = trendsCache
		trendsCacheLock.RUnlock()
		log.Println("Using cached trends data")
	} else {
		// Fetch fresh data
		trends = fetchTrends()

		// Update cache
		trendsCacheLock.Lock()
		trendsCache = trends
		lastUpdated = time.Now()
		trendsCacheLock.Unlock()
		log.Println("Updated trends cache with fresh data")
	}

	// Apply limit
	if len(trends) > limit {
		trends = trends[:limit]
	}

	// Check if we should include user recommendations (for home feed)
	includeRecommendations := c.Query("include_recommendations") == "true"

	// Initialize the response object
	response := gin.H{
		"trends": trends,
	}

	// Include user recommendations if requested and user is authenticated
	if includeRecommendations && userIDStr != "" {
		// Get user recommendations limit, default to 3
		recLimitStr := c.DefaultQuery("rec_limit", "3")
		recLimit, err := strconv.Atoi(recLimitStr)
		if err != nil || recLimit < 1 {
			recLimit = 3
		}

		// Check if the user service client is initialized
		if userServiceClient != nil {
			// Get user recommendations
			users, err := userServiceClient.GetUserRecommendations(userIDStr, recLimit)
			if err == nil && len(users) > 0 {
				// Create response format for recommendations
				var recommendations []gin.H
				for _, user := range users {
					recommendations = append(recommendations, gin.H{
						"id":                  user.ID,
						"username":            user.Username,
						"name":                user.Name,
						"profile_picture_url": user.ProfilePictureURL,
						"bio":                 user.Bio,
						"is_verified":         user.IsVerified,
						"follower_count":      user.FollowerCount,
					})
				}

				// Add recommendations to the response
				response["recommended_users"] = recommendations
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
