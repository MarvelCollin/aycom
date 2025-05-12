package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PredictCategory(c *gin.Context) {
	aiServiceAddr := AppConfig.Services.AIService

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var requestBody map[string]string
	if err := json.Unmarshal(body, &requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if _, exists := requestBody["content"]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'content' field"})
		return
	}

	url := fmt.Sprintf("http://%s/predict/category", aiServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read AI service response"})
		return
	}

	var aiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &aiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse AI service response",
		})
		return
	}

	c.JSON(resp.StatusCode, aiResponse)
}
