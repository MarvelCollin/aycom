package utils

import (
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ensureCORSHeaders ensures CORS headers are set for the response
func ensureCORSHeaders(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000" // Default origin for frontend
	}

	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token")
}

func SendErrorResponse(c *gin.Context, status int, code, message string) {
	// First ensure CORS headers are set
	ensureCORSHeaders(c)

	// Log the error for debugging
	log.Printf("Sending error response: Status=%d, Code=%s, Message=%s", status, code, message)

	// Set content type and send response
	c.Header("Content-Type", "application/json")
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}

func SendValidationErrorResponse(c *gin.Context, fieldErrors map[string]string) {
	// First ensure CORS headers are set
	ensureCORSHeaders(c)

	// Log the validation errors
	log.Printf("Validation error: %v", fieldErrors)

	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "Validation failed",
			"fields":  fieldErrors,
		},
		"validation_errors": fieldErrors,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusBadRequest, response)
}

func SendSuccessResponse(c *gin.Context, status int, data interface{}) {
	// First ensure CORS headers are set
	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// SendDirectSuccessResponse sends a success response without wrapping data in another layer
func SendDirectSuccessResponse(c *gin.Context, status int, data interface{}) {
	// First ensure CORS headers are set
	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")

	// Add success field to data if it's a map
	if dataMap, ok := data.(gin.H); ok {
		dataMap["success"] = true
		c.JSON(status, dataMap)
	} else {
		// If data is not a map, wrap it with success field
		c.JSON(status, gin.H{
			"success": true,
			"data":    data,
		})
	}
}

type PaginationData struct {
	TotalCount  int64 `json:"total_count"`
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	HasMore     bool  `json:"has_more"`
	TotalPages  int   `json:"total_pages"`
}

func CreatePaginationData(totalCount int64, currentPage, perPage int) PaginationData {
	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
	hasMore := currentPage < totalPages

	return PaginationData{
		TotalCount:  totalCount,
		CurrentPage: currentPage,
		PerPage:     perPage,
		HasMore:     hasMore,
		TotalPages:  totalPages,
	}
}

func SendPaginatedResponse(c *gin.Context, status int, items interface{}, pagination PaginationData) {
	// First ensure CORS headers are set
	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")
	c.JSON(status, gin.H{
		"success":    true,
		"data":       items,
		"pagination": pagination,
	})
}
