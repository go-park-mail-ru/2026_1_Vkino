package usecase

import (
	"context"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) SearchMovies(ctx context.Context, query string) (domain.SearchResponse, error) {
	normalized := strings.TrimSpace(query)
	if !domain.ValidateSearchQuery(normalized) {
		return domain.SearchResponse{}, domain.ErrInvalidSearchQuery
	}

	movies, err := u.movieRepo.SearchMovies(ctx, normalized)
	if err != nil {
		return domain.SearchResponse{}, err
	}

	actors, err := u.movieRepo.SearchActors(ctx, normalized)
	if err != nil {
		return domain.SearchResponse{}, err
	}

	resultMovies, err := u.buildMovieCardResponses(ctx, movies)
	if err != nil {
		return domain.SearchResponse{}, err
	}

	resultActors, err := u.buildActorShortResponses(ctx, actors)
	if err != nil {
		return domain.SearchResponse{}, err
	}

	return domain.SearchResponse{Movies: resultMovies, Actors: resultActors}, nil
}
