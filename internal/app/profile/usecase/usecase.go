package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
)

type Usecase interface {
	GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error)
}
