package validatex

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	if len(email) == 0 || len(email) > 64 || !emailRegex.MatchString(email) {
		return false
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return false
	}

	domainPart := parts[1]

	return domainPart != "" &&
		!strings.Contains(domainPart, "..") &&
		strings.Contains(domainPart, ".")
}
