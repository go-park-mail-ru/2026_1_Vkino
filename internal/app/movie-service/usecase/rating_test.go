package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	repomocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestBuildMovieResponse_IncludesExternalRatings(t *testing.T) {
	t.Parallel()

	usecase := NewMovieUsecase(nil, stubFileStorage{}, stubFileStorage{}, stubFileStorage{}, stubFileStorage{})
	reviewRating := 9.0

	resp, err := usecase.buildMovieResponse(context.Background(), &domain.Movie{
		ID:              1,
		Title:           "Interstellar",
		ContentType:     "film",
		ReleaseYear:     2014,
		DurationSeconds: 169 * 60,
		PictureFileKey:  "cards/interstellar.jpg",
		PosterFileKey:   "posters/interstellar.jpg",
		ExternalRatings: []domain.ExternalRating{
			{Source: "IMDb", Value: 8.7, Scale: 10},
			{Source: "Kinopoisk", Value: 8.6, Scale: 10},
		},
		Reviews: []domain.MovieReview{
			{
				ID:             10,
				AuthorUserID:   77,
				AuthorEmail:    "alice@example.com",
				Rating:         &reviewRating,
				Comment:        "Great movie",
				LikesCount:     5,
				DislikesCount:  1,
				ViewerReaction: "like",
			},
		},
	})
	if err != nil {
		t.Fatalf("buildMovieResponse() error = %v", err)
	}

	if len(resp.ExternalRatings) != 2 {
		t.Fatalf("external ratings len = %d, want 2", len(resp.ExternalRatings))
	}

	if resp.ExternalRatings[0].Source != "IMDb" || resp.ExternalRatings[0].Value != 8.7 || resp.ExternalRatings[0].Scale != 10 {
		t.Fatalf("unexpected first external rating: %+v", resp.ExternalRatings[0])
	}

	if len(resp.Reviews) != 1 {
		t.Fatalf("reviews len = %d, want 1", len(resp.Reviews))
	}

	if resp.Reviews[0].Author != "al***@example.com" {
		t.Fatalf("masked author = %q, want %q", resp.Reviews[0].Author, "al***@example.com")
	}

	if resp.Reviews[0].ViewerReaction != "like" {
		t.Fatalf("viewer reaction = %q, want like", resp.Reviews[0].ViewerReaction)
	}
}

func TestGetSelectionByTitle_ReturnsComputedRating(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockMovieRepo(ctrl)
	usecase := NewMovieUsecase(repo, nil, stubFileStorage{}, nil, nil)

	rating := 8.25

	repo.EXPECT().
		GetSelectionByTitle(gomock.Any(), "Top").
		Return(domain.Selection{
			Title:  "Top",
			Rating: &rating,
			Movies: []domain.MovieCard{
				{ID: 1, Title: "Movie", PictureFileKey: "cards/movie.jpg"},
			},
		}, nil)

	resp, err := usecase.GetSelectionByTitle(context.Background(), "Top")
	if err != nil {
		t.Fatalf("GetSelectionByTitle() error = %v", err)
	}

	if resp.Rating == nil || *resp.Rating != rating {
		t.Fatalf("selection rating = %v, want %v", resp.Rating, rating)
	}

	if len(resp.Movies) != 1 || resp.Movies[0].PictureFileKey != "https://cdn.test/cards/movie.jpg" {
		t.Fatalf("unexpected movies payload: %+v", resp.Movies)
	}
}
