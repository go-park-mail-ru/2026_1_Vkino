package domain

import "testing"

func TestDTOName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		named interface {
			Name() string
		}
		want string
	}{
		{
			name:  "selection response",
			named: &SelectionResponse{},
			want:  "selections",
		},
		{
			name:  "movie response",
			named: &MovieResponse{},
			want:  "movies",
		},
		{
			name:  "actor response",
			named: &ActorResponse{},
			want:  "actors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.named.Name(); got != tt.want {
				t.Fatalf("expected name %q, got %q", tt.want, got)
			}
		})
	}
}
