package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/repository"
)

type ProfileUsecase struct {
	userRepo repository.UserRepo
}

func NewProfileUsecase(userRepo repository.UserRepo) *ProfileUsecase {
	return &ProfileUsecase{
		userRepo: userRepo,
	}
}

func (u *ProfileUsecase) GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error) {
	profile, err := u.userRepo.GetProfileByID(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, err
	}

	return profile, nil
}
