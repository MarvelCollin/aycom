package handlers

import (
	"aycom/backend/api-gateway/utils"
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
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Failed to read request body")
		return
	}

	var requestBody map[string]string
	if err := json.Unmarshal(body, &requestBody); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request format")
		return
	}

	if _, exists := requestBody["content"]; !exists {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Missing 'content' field")
		return
	}

	url := fmt.Sprintf("http://%s/predict/category", aiServiceAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
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

	var aiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &aiResponse); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to parse AI service response")
		return
	}

	utils.SendSuccessResponse(c, resp.StatusCode, aiResponse)
}