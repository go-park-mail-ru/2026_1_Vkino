package domain

import (
	"strings"
	"testing"
)

func TestValidateSearchQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "empty", query: "   ", want: false},
		{name: "three characters", query: "mat", want: true},
		{name: "four characters", query: "matr", want: true},
		{name: "no searchable characters", query: "!!!!", want: false},
		{name: "too long", query: strings.Repeat("a", 256), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ValidateSearchQuery(tt.query)
			if got != tt.want {
				t.Fatalf("ValidateSearchQuery(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
