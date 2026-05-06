package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"google.golang.org/grpc"
)

type stubMovieClient struct {
	moviev1.MovieServiceClient
	getAllGenres func(context.Context, *moviev1.GetAllGenresRequest,
		...grpc.CallOption) (*moviev1.GetAllGenresResponse, error)
}

func (s stubMovieClient) GetAllGenres(
	ctx context.Context,
	in *moviev1.GetAllGenresRequest,
	opts ...grpc.CallOption,
) (*moviev1.GetAllGenresResponse, error) {
	return s.getAllGenres(ctx, in, opts...)
}

func TestResolveGenreID(t *testing.T) {
	t.Parallel()

	client := stubMovieClient{
		getAllGenres: func(context.Context, *moviev1.GetAllGenresRequest,
			...grpc.CallOption) (*moviev1.GetAllGenresResponse, error) {
			return &moviev1.GetAllGenresResponse{
				Genres: []*moviev1.GenreShort{
					{Id: 1, Title: "Комедия"},
					{Id: 2, Title: "Драма"},
				},
			}, nil
		},
	}

	tests := []struct {
		name   string
		raw    string
		wantID int64
		wantOK bool
	}{
		{name: "numeric id", raw: "12", wantID: 12, wantOK: true},
		{name: "genre title", raw: "Драма", wantID: 2, wantOK: true},
		{name: "trimmed title", raw: "  Комедия  ", wantID: 1, wantOK: true},
		{name: "missing title", raw: "Триллер", wantID: 0, wantOK: false},
		{name: "empty", raw: " ", wantID: 0, wantOK: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotID, gotOK, err := resolveGenreID(context.Background(), client, tt.raw)
			if err != nil {
				t.Fatalf("resolveGenreID() error = %v", err)
			}

			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("resolveGenreID() = (%d, %v), want (%d, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestParseInt32Query(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		query        string
		defaultValue int32
		want         int32
	}{
		{name: "missing", query: "", defaultValue: 10, want: 10},
		{name: "invalid", query: "limit=bad", defaultValue: 5, want: 5},
		{name: "valid", query: "limit=7", defaultValue: 1, want: 7},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			got := parseInt32Query(req, "limit", tt.defaultValue)

			if got != tt.want {
				t.Fatalf("parseInt32Query() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOrderMovieCardsByIDOrder(t *testing.T) {
	t.Parallel()

	movies := []*moviev1.MovieCard{
		{Id: 3, Title: "Three"},
		nil,
		{Id: 1, Title: "One"},
	}

	ordered := orderMovieCardsByIDOrder([]int64{1, 2, 3}, movies)
	if len(ordered) != 2 {
		t.Fatalf("ordered length = %d, want 2", len(ordered))
	}

	if ordered[0].GetId() != 1 || ordered[1].GetId() != 3 {
		t.Fatalf("ordered ids = (%d, %d), want (1, 3)", ordered[0].GetId(), ordered[1].GetId())
	}
}
