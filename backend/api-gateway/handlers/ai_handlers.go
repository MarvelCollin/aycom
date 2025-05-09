package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PredictCategory godoc
// @Summary Predict category for thread content
// @Description Uses AI to predict the most suitable category for thread content
// @Tags AI
// @Accept json
// @Produce json
// @Param content body map[string]string true "Content to categorize"
// @Success 200 {object} map[string]interface{} "Category prediction response"
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /ai/predict-category [post]
func PredictCategory(c *gin.Context) {
	// Parse the AI service address
	aiServiceAddr := AppConfig.Services.AIService

	// Read request body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Validate the request body
	var requestBody map[string]string
	if err := json.Unmarshal(body, &requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if content field exists
	if _, exists := requestBody["content"]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'content' field"})
		return
	}

	// Forward request to AI service
	url := fmt.Sprintf("http://%s/predict/category", aiServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Read response from AI service
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read AI service response"})
		return
	}

	// Parse JSON response
	var aiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &aiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse AI service response",
		})
		return
	}

	// Return the AI service response
	c.JSON(resp.StatusCode, aiResponse)
}
