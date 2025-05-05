package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	threadProto "aycom/backend/services/thread/proto"

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
			ID:        hashtag.Id,
			Category:  "Hashtag",
			Title:     "#" + hashtag.Text,
			PostCount: int(hashtag.ThreadCount),
		})
	}

	return trends
}

// @Summary Get trends
// @Description Returns trending topics
// @Tags Trends
// @Produce json
// @Router /api/v1/trends [get]
func GetTrends(c *gin.Context) {
	// Check if the route was accessed through the authenticated route
	// If yes, validate the authentication, otherwise continue
	if userID, exists := c.Get("userId"); !exists {
		// If the userId doesn't exist in context, we're likely in a public endpoint
		// Just log the info and continue with the request
		log.Println("Trends accessed via public route")
	} else {
		log.Printf("Trends accessed by authenticated user: %v", userID)
	}

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

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
	})
}
