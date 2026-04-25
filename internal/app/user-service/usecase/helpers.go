package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) profileResponse(ctx context.Context, user *domain2.User) (domain2.ProfileResponse, error) {
	resp := domain2.ProfileResponse{
		Email: user.Email,
		Role:  user.Role,
	}

	if user.Birthdate != nil {
		formatted := user.Birthdate.Format("2006-01-02")
		resp.Birthdate = &formatted
	}

	avatarKey := stringValue(user.AvatarFileKey)
	if u.avatarStore != nil && avatarKey != "" {
		url, err := u.avatarStore.PresignGetObject(ctx, avatarKey, 0)
		if err != nil {
			return domain2.ProfileResponse{}, fmt.Errorf("%w: presign avatar key=%q: %v", domain2.ErrInternal, avatarKey, err)
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
