package postgres

import (
	"strings"
	"unicode"
)

func buildPrefixSearchQuery(query string) string {
	terms := make([]string, 0)
	var term strings.Builder

	flushTerm := func() {
		if term.Len() == 0 {
			return
		}

		terms = append(terms, "'"+term.String()+"':*")
		term.Reset()
	}

	for _, r := range strings.ToLower(query) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			term.WriteRune(r)

			continue
		}

		flushTerm()
	}

	flushTerm()

	return strings.Join(terms, " & ")
}
