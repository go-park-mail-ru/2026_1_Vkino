package validatex

import (
	"regexp"
	"strings"
)

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