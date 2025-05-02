package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Trend struct {
	ID        string `json:"id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	PostCount int    `json:"post_count"`
}

func GetTrends(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

	trends := []Trend{
		{
			ID:        "1",
			Category:  "Technology",
			Title:     "Web Development",
			PostCount: 1500,
		},
		{
			ID:        "2",
			Category:  "Business",
			Title:     "Startup Funding",
			PostCount: 1200,
		},
		{
			ID:        "3",
			Category:  "Science",
			Title:     "AI Research",
			PostCount: 950,
		},
		{
			ID:        "4",
			Category:  "Health",
			Title:     "Wellness",
			PostCount: 800,
		},
		{
			ID:        "5",
			Category:  "Entertainment",
			Title:     "Gaming",
			PostCount: 750,
		},
	}

	// Apply limit
	if len(trends) > limit {
		trends = trends[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
	})
}
