package domain

import (
	"regexp"
	"strings"
	"time"
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

func ValidateEmailQuery(query string) bool {
	return validateEmailQuery(query)
}

func validateEmailQuery(query string) bool {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" || len(trimmedQuery) > 64 {
		return false
	}

	return !strings.ContainsAny(trimmedQuery, " \t\n\r")
}

func ValidateEmail(email string) bool {
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
