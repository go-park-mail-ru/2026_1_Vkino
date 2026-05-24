package usecase

import "testing"

func TestLocalizeMovieContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: contentTypeFilm, in: contentTypeFilm, want: "Фильм"},
		{name: contentTypeSeries, in: contentTypeSeries, want: "Сериал"},
		{name: "unknown", in: "anime", want: "anime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := localizeMovieContentType(tt.in); got != tt.want {
				t.Fatalf("localizeMovieContentType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
