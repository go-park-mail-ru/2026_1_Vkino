package usecase

import (
	"context"
	"fmt"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) profileResponse(ctx context.Context, user *domain.User) (domain.ProfileResponse, error) {
	resp := domain.ProfileResponse{
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
			return domain.ProfileResponse{}, fmt.Errorf("%w: presign avatar key=%q: %w", domain.ErrInternal, avatarKey, err)
		}

		resp.AvatarURL = url
	}

	return resp, nil
}

func (u *UserUsecase) enrichUserSearchAvatarKeys(ctx context.Context, users []domain.UserSearchResult) {
	if u.avatarStore == nil {
		return
	}

	for i := range users {
		key := users[i].AvatarURL
		if key == "" {
			continue
		}

		url, err := u.avatarStore.PresignGetObject(ctx, key, 0)
		if err != nil {
			continue
		}

		users[i].AvatarURL = url
	}
}

func (u *UserUsecase) friendAvatarPresignedURL(ctx context.Context, friend *domain.User) string {
	key := stringValue(friend.AvatarFileKey)
	if key == "" || u.avatarStore == nil {
		return ""
	}

	url, err := u.avatarStore.PresignGetObject(ctx, key, 0)
	if err != nil {
		return ""
	}

	return url
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
