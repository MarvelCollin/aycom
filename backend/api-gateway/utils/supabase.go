package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
)

var (
	supabaseURL   = os.Getenv("SUPABASE_URL")
	supabaseKey   = os.Getenv("SUPABASE_ANON_KEY")
	storageClient *storage_go.Client
)

// InitSupabase initializes the Supabase storage client
func InitSupabase() error {
	if supabaseURL == "" {
		supabaseURL = "https://sdhtnvlmuywinhcglfsu.supabase.co"
	}
	if supabaseKey == "" {
		supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M"
	}

	// Create the Supabase storage client
	storageEndpoint := fmt.Sprintf("%s/storage/v1", strings.TrimRight(supabaseURL, "/"))
	storageClient = storage_go.NewClient(storageEndpoint, supabaseKey, nil)

	// Make sure buckets exist
	ensureBucket("media")
	ensureBucket("thread-media")
	ensureBucket("user-media")
	ensureBucket("profiles")
	ensureBucket("banners")

	return nil
}

// ensureBucket makes sure a bucket exists, creating it if needed
func ensureBucket(bucketName string) {
	// Try to get the bucket first
	_, err := storageClient.GetBucket(bucketName)
	if err != nil {
		// If bucket not found, create it
		_, err = storageClient.CreateBucket(bucketName, storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to create bucket %s: %v\n", bucketName, err)
		} else {
			fmt.Printf("Created bucket: %s\n", bucketName)
		}
	}
}

// UploadFile uploads a file to Supabase storage
func UploadFile(file io.Reader, fileName string, bucket string, folder string) (string, error) {
	// Ensure client is initialized
	if storageClient == nil {
		if err := InitSupabase(); err != nil {
			return "", err
		}
	}

	// Generate a unique file name to avoid collisions
	fileExt := filepath.Ext(fileName)
	baseFilename := strings.TrimSuffix(filepath.Base(fileName), fileExt)
	uniqueID := uuid.New().String()
	safeBaseName := sanitizeFilename(baseFilename)

	// Format: folder/baseFilename-uniqueID.ext
	filePath := ""
	if folder != "" {
		filePath = fmt.Sprintf("%s/%s-%s%s", folder, safeBaseName, uniqueID, fileExt)
	} else {
		filePath = fmt.Sprintf("%s-%s%s", safeBaseName, uniqueID, fileExt)
	}

	// Upload the file to Supabase storage
	_, err := storageClient.UploadFile(bucket, filePath, file)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Get public URL
	urlResult := storageClient.GetPublicUrl(bucket, filePath)
	return urlResult.SignedURL, nil
}

// UploadProfilePicture uploads a profile picture and returns the public URL
func UploadProfilePicture(file io.Reader, fileName string, userID string) (string, error) {
	// Use a dedicated bucket for profile pictures
	bucket := "profiles"
	// Generate a folder structure based on the user ID
	folder := userID[:2] // Using first 2 chars of UUID for folder structure

	return UploadFile(file, fileName, bucket, folder)
}

// UploadBanner uploads a user banner image and returns the public URL
func UploadBanner(file io.Reader, fileName string, userID string) (string, error) {
	// Use a dedicated bucket for banner images
	bucket := "banners"
	// Generate a folder structure based on the user ID
	folder := userID[:2] // Using first 2 chars of UUID for folder structure

	return UploadFile(file, fileName, bucket, folder)
}

// DeleteFile deletes a file from Supabase storage
func DeleteFile(bucket string, filePath string) error {
	// Ensure client is initialized
	if storageClient == nil {
		if err := InitSupabase(); err != nil {
			return err
		}
	}

	_, err := storageClient.RemoveFile(bucket, []string{filePath})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// ExtractFilePathFromURL extracts the file path from a Supabase URL for deletion
func ExtractFilePathFromURL(url, bucket string) string {
	// Example URL: https://sdhtnvlmuywinhcglfsu.supabase.co/storage/v1/object/public/profiles/ab/username-uuid.jpg
	parts := strings.Split(url, fmt.Sprintf("public/%s/", bucket))
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// sanitizeFilename replaces invalid characters in a filename
func sanitizeFilename(filename string) string {
	// Replace characters that might cause issues
	replacer := strings.NewReplacer(
		" ", "-",
		"&", "-and-",
		"=", "-eq-",
		"#", "-hash-",
		"+", "-plus-",
		"@", "-at-",
		"$", "-dollar-",
		"%", "-percent-",
		"?", "",
		"!", "",
		":", "",
		";", "",
		",", "",
		"'", "",
		"\"", "",
		"\\", "",
		"/", "",
		"*", "",
		"|", "",
		"<", "",
		">", "",
	)
	return replacer.Replace(filename)
}
