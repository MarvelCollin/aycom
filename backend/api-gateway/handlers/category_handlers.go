package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	aiServiceAddr := AppConfig.Services.AIService

	url := fmt.Sprintf("http://%s/categories", aiServiceAddr)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "AI service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to read AI service response",
		})
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
			{"id": "travel", "name": "Travel"},
			{"id": "other", "name": "Other"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
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
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"categories": defaultCategories,
		})
		return
	}

	c.JSON(http.StatusOK, aiResponse)
}
