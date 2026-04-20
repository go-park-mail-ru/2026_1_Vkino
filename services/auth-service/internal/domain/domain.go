package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode"
)

type User struct {
	ID               int64
	Email            string
	Password         string
	Birthdate        *time.Time
	AvatarFileKey    *string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func Validate(email, password string) bool {
	return validateEmail(email) && validatePassword(password)
}

func ValidatePassword(password string) bool {
	return validatePassword(password)
}

func validateEmail(email string) bool {
	if len(email) == 0 || len(email) > 64 {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domainPart := parts[1]
	if strings.Contains(domainPart, "..") {
		return false
	}

	if !strings.Contains(domainPart, ".") {
		return false
	}

	return true
}

func validatePassword(password string) bool {
	if len(password) < 6 || len(password) >= 255 {
		return false
	}

	if strings.Contains(password, " ") {
		return false
	}

	var hasLetter bool
	var hasDigit bool

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}