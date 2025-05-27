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

func UploadMedia(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not authenticated"})
		return
	}

	_, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid user ID format"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file provided"})
		return
	}

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

	mediaType := "image"
	if fileExt == ".gif" {
		mediaType = "gif"
	} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" {
		mediaType = "video"
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	bucket := "media"
	folder := mediaType + "s"

	url, err := utils.UploadFile(fileContent, file.Filename, bucket, folder)
	if err != nil {
		log.Printf("Failed to upload file to Supabase: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to upload file: %v", err)})
		return
	}

	mediaID := uuid.New().String()

	thumbnailUrl := ""
	if mediaType == "video" {

	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"id":        mediaID,
		"type":      mediaType,
		"url":       url,
		"thumbnail": thumbnailUrl,
	})
}

func SearchMedia(c *gin.Context) {
	query := c.Query("q")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Media search functionality",
		"query":   query,
		"results": []gin.H{}, 
	})
}