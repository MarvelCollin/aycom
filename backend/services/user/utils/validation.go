package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrNameRequired     = errors.New("name is required")
	ErrNameTooShort     = errors.New("name must be more than 4 characters")
	ErrNameTooLong      = errors.New("name cannot exceed 50 characters")
	ErrNameInvalidChars = errors.New("name must not contain symbols or numbers")

	ErrUsernameRequired     = errors.New("username is required")
	ErrUsernameTooShort     = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong      = errors.New("username cannot exceed 15 characters")
	ErrUsernameInvalidChars = errors.New("username can only contain letters, numbers, and underscores")

	ErrEmailRequired = errors.New("email is required")
	ErrInvalidEmail  = errors.New("invalid email format")

	ErrPasswordRequired  = errors.New("password is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordNoUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber  = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")

	ErrGenderRequired = errors.New("gender is required")
	ErrInvalidGender  = errors.New("gender must be either 'male' or 'female'")

	ErrDOBRequired  = errors.New("date of birth is required")
	ErrInvalidDOB   = errors.New("invalid date of birth format")
	ErrUserTooYoung = errors.New("user must be at least 13 years old")

	ErrSecurityQuestionRequired = errors.New("security question is required")
	ErrSecurityAnswerRequired   = errors.New("security answer is required")
	ErrSecurityAnswerTooShort   = errors.New("security answer must be at least 3 characters")
)

func ValidateName(name string) error {
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)

	if name == "" {
		return ErrNameRequired
	} else if len(name) <= 4 {
		return ErrNameTooShort
	} else if len(name) > 50 {
		return ErrNameTooLong
	} else if !nameRegex.MatchString(name) {
		return ErrNameInvalidChars
	}

	return nil
}

func ValidateUsername(username string) error {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

	if username == "" {
		return ErrUsernameRequired
	} else if len(username) < 3 {
		return ErrUsernameTooShort
	} else if len(username) > 15 {
		return ErrUsernameTooLong
	} else if !usernameRegex.MatchString(username) {
		return ErrUsernameInvalidChars
	}

	return nil
}

func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$`)

	if email == "" {
		return ErrEmailRequired
	} else if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

func ValidatePassword(password string) []error {
	var errors []error

	if password == "" {
		return []error{ErrPasswordRequired}
	}

	if len(password) < 8 {
		errors = append(errors, ErrPasswordTooShort)
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, ErrPasswordNoUpper)
	}

	if !hasLower {
		errors = append(errors, ErrPasswordNoLower)
	}

	if !hasNumber {
		errors = append(errors, ErrPasswordNoNumber)
	}

	if !hasSpecial {
		errors = append(errors, ErrPasswordNoSpecial)
	}

	return errors
}

func ValidateGender(gender string) error {
	if gender == "" {
		return ErrGenderRequired
	}

	gender = strings.ToLower(gender)
	if gender != "male" && gender != "female" {
		return ErrInvalidGender
	}

	return nil
}

func ParseCustomDateFormat(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, ErrDOBRequired
	}

	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return time.Time{}, ErrInvalidDOB
	}

	monthIdx, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, ErrInvalidDOB
	}
	month := monthIdx + 1

	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, ErrInvalidDOB
	}

	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, ErrInvalidDOB
	}

	formattedDateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	return time.Parse("2006-01-02", formattedDateStr)
}

func ValidateDateOfBirth(dateStr string) error {
	if dateStr == "" {
		return ErrDOBRequired
	}

	dob, err := ParseCustomDateFormat(dateStr)
	if err != nil {
		return err
	}

	today := time.Now()
	age := today.Year() - dob.Year()

	if today.Month() < dob.Month() || (today.Month() == dob.Month() && today.Day() < dob.Day()) {
		age--
	}

	if age < 13 {
		return ErrUserTooYoung
	}

	return nil
}

func ValidateSecurityQuestion(question, answer string) error {
	if question == "" {
		return ErrSecurityQuestionRequired
	}

	if answer == "" {
		return ErrSecurityAnswerRequired
	}

	if len(answer) < 3 {
		return ErrSecurityAnswerTooShort
	}

	return nil
}
