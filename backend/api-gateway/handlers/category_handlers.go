package handlers

import (
	"aycom/backend/api-gateway/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "thread_categories"

	// Try to get from cache first
	var cachedResponse map[string]interface{}
	if err := utils.GetCache(ctx, cacheKey, &cachedResponse); err == nil {
		c.Header("X-Cache", "HIT")
		utils.SendSuccessResponse(c, http.StatusOK, cachedResponse)
		return
	}

	// Cache miss - fetch from AI service
	c.Header("X-Cache", "MISS")
	aiServiceAddr := AppConfig.Services.AIService

	url := fmt.Sprintf("http://%s/categories", aiServiceAddr)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "AI service unavailable")
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to read AI service response")
		return
	}

	if resp.StatusCode != http.StatusOK {
		defaultCategories := []map[string]interface{}{
			{"id": "technology", "name": "Technology"},
			{"id": "health", "name": "Health"},
			{"id": "education", "name": "Education"},
			{"id": "entertainment", "name": "Entertainment"},
			{"id": "science", "name": "Science"},
			{"id": "sports", "name": "Sports"},
			{"id": "politics", "name": "Politics"},
			{"id": "business", "name": "Business"},
			{"id": "lifestyle", "name": "Lifestyle"},
			{"id": "travel", "name": "Travel"}, {"id": "other", "name": "Other"},
		}

		// Cache the default response
		_ = utils.SetCache(ctx, cacheKey, gin.H{"categories": defaultCategories}, 24*time.Hour)

		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"categories": defaultCategories,
		})
		return
	}

	var aiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &aiResponse); err != nil {
		defaultCategories := []map[string]interface{}{
			{"id": "technology", "name": "Technology"},
			{"id": "health", "name": "Health"},
			{"id": "education", "name": "Education"},
			{"id": "entertainment", "name": "Entertainment"},
			{"id": "science", "name": "Science"},
			{"id": "sports", "name": "Sports"},
			{"id": "politics", "name": "Politics"},
			{"id": "business", "name": "Business"},
			{"id": "lifestyle", "name": "Lifestyle"},
			{"id": "travel", "name": "Travel"},
			{"id": "other", "name": "Other"},
		}

		// Cache the default response
		_ = utils.SetCache(ctx, cacheKey, gin.H{"categories": defaultCategories}, 24*time.Hour)

		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"categories": defaultCategories,
		})
		return
	}

	// Cache the AI service response
	_ = utils.SetCache(ctx, cacheKey, aiResponse, 24*time.Hour)

	utils.SendSuccessResponse(c, http.StatusOK, aiResponse)
}
