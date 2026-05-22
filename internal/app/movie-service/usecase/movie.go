package usecase

import (
	"context"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

func (u *MovieUsecase) GetMovieByID(ctx context.Context, movieID int64) (domain.MovieResponse, error) {
	if movieID <= 0 {
		return domain.MovieResponse{}, domain.ErrInvalidMovieID
	}

	movie, err := u.movieRepo.GetMovieByID(ctx, movieID)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	viewerUserID := viewerIDFromContext(ctx)

	movie.Reviews, err = u.movieRepo.GetMovieReviews(ctx, movieID, viewerUserID)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp, err := u.buildMovieResponse(ctx, movie)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	u.applyFavoriteFlag(ctx, &resp, viewerUserID, movieID)

	return resp, nil
}

func viewerIDFromContext(ctx context.Context) int64 {
	authCtx, err := authctx.FromContext(ctx)
	if err != nil {
		return 0
	}

	return authCtx.UserID
}

func (u *MovieUsecase) applyFavoriteFlag(
	ctx context.Context,
	resp *domain.MovieResponse,
	viewerUserID, movieID int64,
) {
	if viewerUserID <= 0 {
		return
	}

	isFavorite, err := u.movieRepo.IsFavorite(ctx, viewerUserID, movieID)
	if err == nil {
		resp.IsFavorite = isFavorite
	}
}

func (u *MovieUsecase) GetMoviesByIDs(ctx context.Context, movieIDs []int64) ([]domain.MovieCardResponse, error) {
	movies, err := u.movieRepo.GetMovieCardsByIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}

	return u.buildMovieCardResponses(ctx, movies)
}
