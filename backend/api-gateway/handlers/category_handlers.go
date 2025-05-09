package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary Get available categories
// @Description Returns a list of available thread categories
// @Tags Categories
// @Produce json
// @Success 200 {object} map[string]interface{} "List of categories"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /categories [get]
func GetCategories(c *gin.Context) {
	// Parse the AI service address
	aiServiceAddr := AppConfig.Services.AIService

	// Create the URL for the AI service endpoint
	url := fmt.Sprintf("http://%s/categories", aiServiceAddr)

	// Create HTTP client with timeout
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Call AI service to get categories
	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "AI service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	// Read response from AI service
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to read AI service response",
		})
		return
	}

	// If AI service doesn't return categories, use default ones
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
			{"id": "travel", "name": "Travel"},
			{"id": "other", "name": "Other"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"categories": defaultCategories,
		})
		return
	}

	// Parse JSON response
	var aiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &aiResponse); err != nil {
		// If parsing fails, return default categories
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
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"categories": defaultCategories,
		})
		return
	}

	// Forward the AI service response
	c.JSON(http.StatusOK, aiResponse)
}
