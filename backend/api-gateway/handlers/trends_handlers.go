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

// fetchTrends retrieves trend data from the database
// Currently a placeholder that should be updated to use real data
func fetchTrends() []Trend {
	// This function should be updated to fetch real trending data
	// from the thread/hashtag service or database
	log.Println("Fetching trends - to be implemented with real data source")

	// Placeholder data until real data implementation is ready
	// Note: This should be replaced with actual database queries
	trends := []Trend{
		{
			ID:        "1",
			Category:  "Technology",
			Title:     "WebDev",
			PostCount: 532,
		},
		{
			ID:        "2",
			Category:  "Business",
			Title:     "Startup",
			PostCount: 435,
		},
		{
			ID:        "3",
			Category:  "Science",
			Title:     "AI",
			PostCount: 378,
		},
		{
			ID:        "4",
			Category:  "Health",
			Title:     "Wellness",
			PostCount: 326,
		},
		{
			ID:        "5",
			Category:  "Tech",
			Title:     "MobileApps",
			PostCount: 289,
		},
	}

	return trends
}

// @Summary Get trends
// @Description Returns trending topics
// @Tags Trends
// @Produce json
// @Router /api/v1/trends [get]
func GetTrends(c *gin.Context) {
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
