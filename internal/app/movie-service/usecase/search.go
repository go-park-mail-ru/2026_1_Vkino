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

	result := domain.SearchResponse{
		Movies: make([]domain.MovieCardResponse, 0, len(movies)),
		Actors: make([]domain.ActorShortResponse, 0, len(actors)),
	}

	for _, movie := range movies {
		card, buildErr := u.buildMovieCardResponse(ctx, movie)
		if buildErr != nil {
			return domain.SearchResponse{}, buildErr
		}

		result.Movies = append(result.Movies, card)
	}

	for _, actor := range actors {
		item, buildErr := u.buildActorShortResponse(ctx, actor)
		if buildErr != nil {
			return domain.SearchResponse{}, buildErr
		}

		result.Actors = append(result.Actors, item)
	}

	return result, nil
}
