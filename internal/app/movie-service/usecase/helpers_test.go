package usecase

import "testing"

func TestLocalizeMovieContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "film", in: "film", want: "Фильм"},
		{name: "series", in: "series", want: "Сериал"},
		{name: "unknown", in: "anime", want: "anime"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := localizeMovieContentType(tt.in); got != tt.want {
				t.Fatalf("localizeMovieContentType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
