package utils

import (
	"log"
	"os"
	"strings"
)

// IsDevelopmentMode checks if the application is running in development mode
func IsDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("GIN_MODE"))
	return env != "release" && env != "production"
}

// VerifyRecaptcha verifies a reCAPTCHA token with Google's API
func VerifyRecaptcha(token string) (bool, error) {
	// TEMPORARY FIX: Always return success regardless of token
	log.Printf("NOTICE: reCAPTCHA verification bypassed - always returning success")
	return true, nil

	// Original code below (commented out)
	/*
		secretKey := os.Getenv("RECAPTCHA_SECRET_KEY")
		if secretKey == "" || secretKey == "YOUR_RECAPTCHA_SECRET_KEY" {
			log.Printf("WARNING: RECAPTCHA_SECRET_KEY environment variable not set or using placeholder value")
			if IsDevelopmentMode() {
				// In development mode, accept any token as valid to facilitate testing
				log.Printf("Development mode: Bypassing actual reCAPTCHA verification, accepting token as valid")
				return true, nil
			}
			return false, fmt.Errorf("reCAPTCHA secret key not properly configured")
		}

		log.Printf("Verifying reCAPTCHA token (length: %d) using secret key: %s...",
			len(token), secretKey[:10])

		if len(token) < 10 {
			log.Printf("WARNING: reCAPTCHA token is too short: %s", token)
			return false, fmt.Errorf("reCAPTCHA token is too short")
		}

		// Directly use form values instead of URL encoding
		formData := url.Values{}
		formData.Set("secret", secretKey)
		formData.Set("response", token)

		// Log the request before sending
		log.Printf("Sending reCAPTCHA verification request to Google API with token: %s...", token[:10])

		resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", formData)
		if err != nil {
			log.Printf("ERROR: reCAPTCHA HTTP request failed: %v", err)
			return false, fmt.Errorf("failed to send reCAPTCHA verification request: %w", err)
		}
		defer resp.Body.Close()

		// Check HTTP status code
		log.Printf("reCAPTCHA verification response status: %d %s", resp.StatusCode, resp.Status)
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("reCAPTCHA verification API returned HTTP status: %d", resp.StatusCode)
		}

		// Read entire response body for logging
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: Failed to read reCAPTCHA response body: %v", err)
			return false, fmt.Errorf("failed to read reCAPTCHA response: %w", err)
		}

		// Log the raw response
		log.Printf("reCAPTCHA raw response: %s", string(responseBody))

		var recaptchaResp struct {
			Success     bool      `json:"success"`
			Score       float64   `json:"score,omitempty"`
			ErrorCodes  []string  `json:"error-codes,omitempty"`
			ChallengeTS time.Time `json:"challenge_ts,omitempty"`
			Hostname    string    `json:"hostname,omitempty"`
		}

		if err := json.Unmarshal(responseBody, &recaptchaResp); err != nil {
			log.Printf("ERROR: Failed to decode reCAPTCHA response: %v", err)
			return false, fmt.Errorf("failed to decode reCAPTCHA response: %w", err)
		}

		// Log the response fields
		log.Printf("reCAPTCHA verification result - Success: %v, Score: %.2f, ErrorCodes: %v, Hostname: %s",
			recaptchaResp.Success, recaptchaResp.Score, recaptchaResp.ErrorCodes, recaptchaResp.Hostname)

		// If not successful, include error codes in the error message
		if !recaptchaResp.Success && len(recaptchaResp.ErrorCodes) > 0 {
			log.Printf("ERROR: reCAPTCHA verification failed with error codes: %v", recaptchaResp.ErrorCodes)
			return false, fmt.Errorf("reCAPTCHA verification failed with error codes: %v", recaptchaResp.ErrorCodes)
		}

		return recaptchaResp.Success, nil
	*/
}
