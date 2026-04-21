package validatex

import (
	"strings"
)

func ValidateEmailQuery(query string) bool {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" || len(trimmedQuery) > 64 {
		return false
	}

	return !strings.ContainsAny(trimmedQuery, " \t\n\r")
}