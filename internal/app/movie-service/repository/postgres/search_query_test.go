package postgres

import "testing"

func TestBuildPrefixSearchQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "single term", query: "Matrix", want: "'matrix':*"},
		{name: "multiple terms", query: "John Wick 4", want: "'john':* & 'wick':* & '4':*"},
		{name: "punctuation separators", query: "spider-man: home", want: "'spider':* & 'man':* & 'home':*"},
		{name: "no terms", query: "!!!", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildPrefixSearchQuery(tt.query)
			if got != tt.want {
				t.Fatalf("buildPrefixSearchQuery(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}
