package utils

import (
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

func SendErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}

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

func SendSuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
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
	c.JSON(status, gin.H{
		"success":    true,
		"data":       items,
		"pagination": pagination,
	})
}