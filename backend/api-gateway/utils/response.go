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

func ensureCORSHeaders(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000" 
	}

	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token")
}

func SendErrorResponse(c *gin.Context, status int, code, message string) {

	ensureCORSHeaders(c)

	log.Printf("Sending error response: Status=%d, Code=%s, Message=%s", status, code, message)

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

	ensureCORSHeaders(c)

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

	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func SendDirectSuccessResponse(c *gin.Context, status int, data interface{}) {

	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")

	if dataMap, ok := data.(gin.H); ok {
		dataMap["success"] = true
		c.JSON(status, dataMap)
	} else {

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

	ensureCORSHeaders(c)

	c.Header("Content-Type", "application/json")
	c.JSON(status, gin.H{
		"success":    true,
		"data":       items,
		"pagination": pagination,
	})
}