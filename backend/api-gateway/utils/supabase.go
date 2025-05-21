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

func InitSupabase() error {
	if supabaseURL == "" {
		supabaseURL = "https://sdhtnvlmuywinhcglfsu.supabase.co"
	}
	if supabaseKey == "" {
		supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNkaHRudmxtdXl3aW5oY2dsZnN1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDU5MDE4NzUsImV4cCI6MjA2MTQ3Nzg3NX0.Jknb2LNtRgma15sEX0sgLHMPegpCQ1f-05QbZEgHq8M"
	}
	storageEndpoint := fmt.Sprintf("%s/storage/v1", strings.TrimRight(supabaseURL, "/"))
	storageClient = storage_go.NewClient(storageEndpoint, supabaseKey, nil)
	ensureBucket("media")
	ensureBucket("thread-media")
	ensureBucket("user-media")
	ensureBucket("profiles")
	ensureBucket("banners")
	return nil
}
func ensureBucket(bucketName string) {
	_, err := storageClient.GetBucket(bucketName)
	if err != nil {
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
func UploadFile(file io.Reader, fileName string, bucket string, folder string) (string, error) {
	if storageClient == nil {
		if err := InitSupabase(); err != nil {
			return "", err
		}
	}
	fileExt := filepath.Ext(fileName)
	baseFilename := strings.TrimSuffix(filepath.Base(fileName), fileExt)
	uniqueID := uuid.New().String()
	safeBaseName := sanitizeFilename(baseFilename)
	filePath := ""
	if folder != "" {
		filePath = fmt.Sprintf("%s/%s-%s%s", folder, safeBaseName, uniqueID, fileExt)
	} else {
		filePath = fmt.Sprintf("%s-%s%s", safeBaseName, uniqueID, fileExt)
	}
	_, err := storageClient.UploadFile(bucket, filePath, file)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	urlResult := storageClient.GetPublicUrl(bucket, filePath)
	return urlResult.SignedURL, nil
}
func UploadProfilePicture(file io.Reader, fileName string, userID string) (string, error) {
	bucket := "profiles"
	folder := userID[:2]
	return UploadFile(file, fileName, bucket, folder)
}
func UploadBanner(file io.Reader, fileName string, userID string) (string, error) {
	bucket := "banners"
	folder := userID[:2]
	return UploadFile(file, fileName, bucket, folder)
}
func DeleteFile(bucket string, filePath string) error {
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
func ExtractFilePathFromURL(url, bucket string) string {
	parts := strings.Split(url, fmt.Sprintf("public/%s/", bucket))
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
func sanitizeFilename(filename string) string {
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
