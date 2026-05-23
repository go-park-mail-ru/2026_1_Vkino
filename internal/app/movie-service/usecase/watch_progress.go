package usecase

import (
	"context"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetContinueWatching(
	ctx context.Context,
	userID int64,
	limit int32,
) ([]domain.WatchProgressItemResponse, error) {
	if userID <= 0 {
		return nil, domain.ErrInternal
	}

	if limit <= 0 {
		limit = 5
	}

	if err := u.ensureSmartContinueAllowed(ctx, userID); err != nil {
		return nil, err
	}

	items, err := u.movieRepo.GetContinueWatching(ctx, userID, limit)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return buildWatchProgressResponses(items), nil
}

func (u *MovieUsecase) GetWatchHistory(
	ctx context.Context,
	userID int64,
	limit int32,
	minProgress float64,
) ([]domain.WatchProgressItemResponse, error) {
	if userID <= 0 {
		return nil, domain.ErrInternal
	}

	if limit <= 0 {
		limit = 10
	}

	items, err := u.movieRepo.GetWatchHistory(ctx, userID, limit, minProgress)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return buildWatchProgressResponses(items), nil
}

func buildWatchProgressResponses(items []domain.WatchProgressItem) []domain.WatchProgressItemResponse {
	resp := make([]domain.WatchProgressItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, domain.WatchProgressItemResponse(item))
	}

	return resp
}
