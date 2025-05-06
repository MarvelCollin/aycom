package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"aycom/backend/api-gateway/utils"

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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not authenticated"})
		return
	}

	// Convert userID to string and validate
	_, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid user ID format"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file provided"})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "File type not allowed"})
		return
	}

	// Determine media type
	mediaType := "image"
	if fileExt == ".gif" {
		mediaType = "gif"
	} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" {
		mediaType = "video"
	}

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	// Determine bucket and folder based on media type
	bucket := "media"
	folder := mediaType + "s"

	// Upload to Supabase
	url, err := utils.UploadFile(fileContent, file.Filename, bucket, folder)
	if err != nil {
		log.Printf("Failed to upload file to Supabase: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to upload file: %v", err)})
		return
	}

	// Generate unique ID for the media
	mediaID := uuid.New().String()

	// For video files, we would need to generate a thumbnail
	// But for now, just leave it empty as we don't have thumbnail generation yet
	thumbnailUrl := ""
	if mediaType == "video" {
		// In a real implementation, we'd generate and upload a thumbnail
		// thumbnailUrl = "https://your-thumbnail-url.com"
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"id":        mediaID,
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
