package usecase

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	repomocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/mocks"
	"go.uber.org/mock/gomock"
)

type stubFileStorage struct{}

func (stubFileStorage) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	return nil
}

func (stubFileStorage) DeleteObject(ctx context.Context, key string) error {
	return nil
}

func (stubFileStorage) PresignGetObject(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return "https://cdn.test/" + key, nil
}

func (stubFileStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func TestSearchMovies_ReturnsMoviesAndActors(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := repomocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, nil, stubFileStorage{}, stubFileStorage{}, nil)

	mr.EXPECT().SearchMovies(gomock.Any(), "matrix").Return([]domain.MovieCard{
		{ID: 1, Title: "The Matrix", PictureFileKey: "movies/matrix.jpg"},
	}, nil)
	mr.EXPECT().SearchActors(gomock.Any(), "matrix").Return([]domain.ActorShort{
		{ID: 7, FullName: "Carrie-Anne Moss", PictureFileKey: "actors/moss.jpg"},
	}, nil)

	result, err := u.SearchMovies(context.Background(), "matrix")
	if err != nil {
		t.Fatalf("SearchMovies: %v", err)
	}

	if len(result.Movies) != 1 {
		t.Fatalf("movies len = %d, want 1", len(result.Movies))
	}

	if got := result.Movies[0].PictureFileKey; got != "https://cdn.test/movies/matrix.jpg" {
		t.Fatalf("movie img_url = %q, want %q", got, "https://cdn.test/movies/matrix.jpg")
	}

	if len(result.Actors) != 1 {
		t.Fatalf("actors len = %d, want 1", len(result.Actors))
	}

	if got := result.Actors[0].PictureFileKey; got != "https://cdn.test/actors/moss.jpg" {
		t.Fatalf("actor img_url = %q, want %q", got, "https://cdn.test/actors/moss.jpg")
	}
}

func TestSearchMovies_InvalidQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	mr := repomocks.NewMockMovieRepo(ctrl)
	u := NewMovieUsecase(mr, nil, nil, nil, nil)

	_, err := u.SearchMovies(context.Background(), "   ")
	if err != domain.ErrInvalidSearchQuery {
		t.Fatalf("SearchMovies error = %v, want %v", err, domain.ErrInvalidSearchQuery)
	}
}
