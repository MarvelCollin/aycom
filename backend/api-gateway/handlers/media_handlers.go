package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"aycom/backend/api-gateway/utils"
)

func UploadMedia(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	_, ok := userID.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid user ID format")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "No file provided")
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
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "File type not allowed")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to open file")
		return
	}
	defer fileContent.Close()

	bucket := "media"
	folder := mediaType + "s"

	url, err := utils.UploadFile(fileContent, file.Filename, bucket, folder)
	if err != nil {
		log.Printf("Failed to upload file to Supabase: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to upload file: %v", err))
		return
	}

	mediaID := uuid.New().String()

	thumbnailUrl := ""
	if mediaType == "video" {

	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"id":        mediaID,
		"type":      mediaType,
		"url":       url,
		"thumbnail": thumbnailUrl,
	})
}

func SearchMedia(c *gin.Context) {
	query := c.Query("q")

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Media search functionality",
		"query":   query,
		"results": []gin.H{},
	})
}