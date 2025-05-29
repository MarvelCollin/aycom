package handlers

import (
	"aycom/backend/api-gateway/utils"
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

var (
	trendsCache     []Trend
	trendsCacheLock sync.RWMutex
	lastUpdated     time.Time
	cacheExpiration = 10 * time.Minute
)

func fetchTrends() []Trend {
	log.Println("Fetching trending hashtags from thread service")

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("Failed to connect to thread service: %v", err)
		log.Println("Returning mock trending hashtags instead")

		return []Trend{
			{
				ID:        "golang",
				Category:  "Hashtag",
				Title:     "#golang",
				PostCount: 5243,
			},
			{
				ID:        "development",
				Category:  "Hashtag",
				Title:     "#development",
				PostCount: 4328,
			},
			{
				ID:        "coding",
				Category:  "Hashtag",
				Title:     "#coding",
				PostCount: 3921,
			},
			{
				ID:        "webdev",
				Category:  "Hashtag",
				Title:     "#webdev",
				PostCount: 3157,
			},
			{
				ID:        "javascript",
				Category:  "Hashtag",
				Title:     "#javascript",
				PostCount: 2845,
			},
		}
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetTrendingHashtags(ctx, &threadProto.GetTrendingHashtagsRequest{
		Limit: 20,
	})
	if err != nil {
		log.Printf("Failed to get trending hashtags: %v", err)
		log.Println("Returning mock trending hashtags instead")

		return []Trend{
			{
				ID:        "golang",
				Category:  "Hashtag",
				Title:     "#golang",
				PostCount: 5243,
			},
			{
				ID:        "development",
				Category:  "Hashtag",
				Title:     "#development",
				PostCount: 4328,
			},
			{
				ID:        "coding",
				Category:  "Hashtag",
				Title:     "#coding",
				PostCount: 3921,
			},
			{
				ID:        "webdev",
				Category:  "Hashtag",
				Title:     "#webdev",
				PostCount: 3157,
			},
			{
				ID:        "javascript",
				Category:  "Hashtag",
				Title:     "#javascript",
				PostCount: 2845,
			},
		}
	}

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

func GetTrends(c *gin.Context) {

	var userIDStr string
	if userID, exists := c.Get("userId"); !exists {

		log.Println("Trends accessed via public route")
	} else {
		log.Printf("Trends accessed by authenticated user: %v", userID)
		userIDStr = userID.(string)
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

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

		trends = fetchTrends()

		trendsCacheLock.Lock()
		trendsCache = trends
		lastUpdated = time.Now()
		trendsCacheLock.Unlock()
		log.Println("Updated trends cache with fresh data")
	}

	if len(trends) > limit {
		trends = trends[:limit]
	}

	includeRecommendations := c.Query("include_recommendations") == "true"

	response := gin.H{
		"trends": trends,
	}

	if includeRecommendations && userIDStr != "" {

		recLimitStr := c.DefaultQuery("rec_limit", "3")
		recLimit, err := strconv.Atoi(recLimitStr)
		if err != nil || recLimit < 1 {
			recLimit = 3
		}

		if userServiceClient != nil {

			users, err := userServiceClient.GetUserRecommendations(userIDStr, recLimit)
			if err == nil && len(users) > 0 {

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

				response["recommended_users"] = recommendations
			}
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, response)
}
