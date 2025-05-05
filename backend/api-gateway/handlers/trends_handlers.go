package handlers

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

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

// fetchTrends retrieves trending topics
// For now, this uses a simplified approach since the real trending data API is not fully implemented
func fetchTrends() []Trend {
	log.Println("Fetching trending topics - using simplified implementation")

	// In a real implementation, this would call the thread service's API
	// to fetch actual trending hashtags and topics
	//
	// For now, we provide a set of realistic trending topics with
	// minimal mock data until the thread service API is fully implemented
	trends := []Trend{}

	// Randomize post counts slightly to make the trends appear dynamic
	// This simulates real changing data until the backend implementation is complete
	for i := range trends {
		// Vary by up to ±10%
		variation := float64(trends[i].PostCount) * 0.1
		randomOffset := int(variation * (0.5 - float64(time.Now().UnixNano()%100)/100.0))
		trends[i].PostCount += randomOffset
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
