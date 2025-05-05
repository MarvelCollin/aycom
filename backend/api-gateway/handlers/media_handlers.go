package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadMedia handles uploading a media file (image, gif, video)
// @Summary Upload media
// @Description Upload a media file (image, gif, video)
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file to upload"
// @Param type formData string false "Media type (image, gif, video)" Enums(image, gif, video)
// @Success 200 {object} models.MediaUploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/media [post]
func UploadMedia(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Check file type
	fileExt := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".mp4":  true,
		".webm": true,
		".mov":  true,
	}

	if !allowedExts[fileExt] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + fileExt

	// TODO: Actually save the file to storage/cloud
	// This is a placeholder implementation

	// Determine media type
	mediaType := "image"
	if fileExt == ".gif" {
		mediaType = "gif"
	} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" {
		mediaType = "video"
	}

	// Placeholder URL - in actual implementation, this would be the URL to the uploaded file
	url := "https://media.example.com/" + filename
	thumbnailUrl := ""
	if mediaType == "video" {
		thumbnailUrl = "https://media.example.com/thumbnails/" + filename + ".jpg"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        filename,
		"type":      mediaType,
		"url":       url,
		"thumbnail": thumbnailUrl,
	})
}

// SearchMedia handles search requests for media content
func SearchMedia(c *gin.Context) {
	query := c.Query("q")

	// For now, just return a simple response
	// In a real implementation, this would query media from a database or service
	c.JSON(200, gin.H{
		"success": true,
		"message": "Media search functionality",
		"query":   query,
		"results": []gin.H{}, // Empty results for now
	})
}
