package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/domain"
)

func (u *UserUsecase) profileResponse(ctx context.Context, user *domain.User) (domain.ProfileResponse, error) {
	resp := domain.ProfileResponse{
		Email: user.Email,
	}

	if user.Birthdate != nil {
		formatted := user.Birthdate.Format("2006-01-02")
		resp.Birthdate = &formatted
	}

	avatarKey := stringValue(user.AvatarFileKey)
	if u.avatarStore != nil && avatarKey != "" {
		url, err := u.avatarStore.PresignGetObject(ctx, avatarKey, 0)
		if err != nil {
			return domain.ProfileResponse{}, fmt.Errorf("%w: presign avatar key=%q: %v", domain.ErrInternal, avatarKey, err)
		}

		resp.AvatarURL = url
	}

	return resp, nil
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
