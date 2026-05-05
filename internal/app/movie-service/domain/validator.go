package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	minSearchQueryRunes = 3
	maxSearchQueryRunes = 255
)

func ValidateSelectionTitle(title string) bool {
	trimmed := strings.TrimSpace(title)

	return trimmed != "" && len(trimmed) <= 255
}

func ValidateSearchQuery(query string) bool {
	trimmed := strings.TrimSpace(query)
	queryLen := utf8.RuneCountInString(trimmed)
	if queryLen < minSearchQueryRunes || queryLen > maxSearchQueryRunes {
		return false
	}

	return strings.IndexFunc(trimmed, func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}) >= 0
}
