package domain

import "strings"

func ValidateSelectionTitle(title string) bool {
	trimmed := strings.TrimSpace(title)
	return trimmed != "" && len(trimmed) <= 255
}

func ValidateSearchQuery(query string) bool {
	trimmed := strings.TrimSpace(query)
	return trimmed != "" && len(trimmed) <= 255
}
