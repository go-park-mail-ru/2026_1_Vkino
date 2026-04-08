package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
)

type UserRepo interface {
	GetProfileByID(ctx context.Context, id int64) (domain.ProfileResponse, error)
}
