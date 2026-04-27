package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetContinueWatching(ctx context.Context, userID int64, limit int32) ([]domain2.WatchProgressItemResponse, error) {
	if userID <= 0 {
		return nil, domain2.ErrInternal
	}
	if limit <= 0 {
		limit = 5
	}
	items, err := u.movieRepo.GetContinueWatching(ctx, userID, limit)
	if err != nil {
		return nil, domain2.ErrInternal
	}
	resp := make([]domain2.WatchProgressItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, domain2.WatchProgressItemResponse(item))
	}
	return resp, nil
}

func (u *MovieUsecase) GetWatchHistory(ctx context.Context, userID int64, limit int32, minProgress float64) ([]domain2.WatchProgressItemResponse, error) {
	if userID <= 0 {
		return nil, domain2.ErrInternal
	}
	if limit <= 0 {
		limit = 10
	}
	items, err := u.movieRepo.GetWatchHistory(ctx, userID, limit, minProgress)
	if err != nil {
		return nil, domain2.ErrInternal
	}
	resp := make([]domain2.WatchProgressItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, domain2.WatchProgressItemResponse(item))
	}
	return resp, nil
}
