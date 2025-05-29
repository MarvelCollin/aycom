package utils

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   ErrorDetails `json:"error"`
}

// ErrorDetails contains specific error information
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}

// SendValidationErrorResponse sends a validation error response with field-specific errors
func SendValidationErrorResponse(c *gin.Context, fieldErrors map[string]string) {
	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "Validation failed",
			"fields":  fieldErrors,
		},
	}
	c.JSON(http.StatusBadRequest, response)
}

// SendSuccessResponse sends a standardized success response
func SendSuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// PaginationData represents standardized pagination metadata
type PaginationData struct {
	TotalCount  int64 `json:"total_count"`
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	HasMore     bool  `json:"has_more"`
	TotalPages  int   `json:"total_pages"`
}

// CreatePaginationData creates standardized pagination metadata
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

// SendPaginatedResponse sends a success response with standardized pagination metadata
func SendPaginatedResponse(c *gin.Context, status int, items interface{}, pagination PaginationData) {
	c.JSON(status, gin.H{
		"success":    true,
		"data":       items,
		"pagination": pagination,
	})
}
