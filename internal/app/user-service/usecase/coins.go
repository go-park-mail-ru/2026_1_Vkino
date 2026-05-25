package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

const (
	defaultCoinsHistoryLimit int32 = 50
	maxCoinsHistoryLimit     int32 = 100
)

func (u *UserUsecase) GetVKinoCoinsHistory(
	ctx context.Context,
	userID int64,
	limit, offset int32,
) (domain.VKinoCoinsHistoryResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.VKinoCoinsHistoryResponse{}, domain.ErrInvalidToken
	}

	if limit <= 0 {
		limit = defaultCoinsHistoryLimit
	}

	if limit > maxCoinsHistoryLimit {
		limit = maxCoinsHistoryLimit
	}

	if offset < 0 {
		offset = 0
	}

	items, totalCount, err := u.userRepo.GetVKinoCoinsHistory(ctx, userID, limit, offset)
	if err != nil {
		return domain.VKinoCoinsHistoryResponse{}, domain.ErrInternal
	}

	return domain.VKinoCoinsHistoryResponse{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}
