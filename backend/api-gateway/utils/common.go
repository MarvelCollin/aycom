package utils

import (
	"log"
	"os"
	"strings"
)

func IsDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("GIN_MODE"))
	return env != "release" && env != "production"
}

func VerifyRecaptcha(token string) (bool, error) {

	log.Printf("NOTICE: reCAPTCHA verification bypassed - always returning success")
	return true, nil

}
