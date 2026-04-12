package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func (u *AuthUsecase) GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, domain.ErrInvalidToken
	}

	return u.profileResponse(ctx, user)
}

func (u *AuthUsecase) UpdateProfile(
	ctx context.Context,
	userID int64,
	birthdate string,
	body io.Reader,
	size int64,
	contentType string,
) (domain.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, domain.ErrInvalidToken
	}

	user, err = u.updateBirthdateIfProvided(ctx, userID, birthdate, user)
	if err != nil {
		return domain.ProfileResponse{}, err
	}

	user, err = u.updateAvatarIfProvided(ctx, userID, user, body, size, contentType)
	if err != nil {
		return domain.ProfileResponse{}, err
	}

	return u.profileResponse(ctx, user)
}

func (u *AuthUsecase) updateBirthdateIfProvided(
	ctx context.Context,
	userID int64,
	birthdate string,
	user *domain.User,
) (*domain.User, error) {
	trimmedBirthdate := strings.TrimSpace(birthdate)
	if trimmedBirthdate == "" {
		return user, nil
	}

	parsedBirthdate, err := parseBirthdate(trimmedBirthdate)
	if err != nil {
		return nil, err
	}

	updatedUser, err := u.userRepo.UpdateBirthdate(ctx, userID, parsedBirthdate)
	if err != nil {
		return nil, fmt.Errorf("%w: update birthdate in repository: %v", domain.ErrInternal, err)
	}

	return updatedUser, nil
}

func parseBirthdate(rawBirthdate string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02", rawBirthdate)
	if err != nil || parsed.After(time.Now()) {
		return nil, domain.ErrInvalidBirthdate
	}

	return &parsed, nil
}

func (u *AuthUsecase) updateAvatarIfProvided(
	ctx context.Context,
	userID int64,
	user *domain.User,
	body io.Reader,
	size int64,
	contentType string,
) (*domain.User, error) {
	if body == nil {
		return user, nil
	}

	if u.avatarStore == nil {
		return nil, fmt.Errorf("%w: avatar storage is not configured", domain.ErrInternal)
	}

	if size <= 0 {
		return nil, domain.ErrInvalidAvatar
	}

	avatarBytes, err := io.ReadAll(body)
	if err != nil || len(avatarBytes) == 0 {
		if err != nil {
			log.Printf("user.update_profile: failed to read avatar body user_id=%d err=%v", userID, err)
		}
		return nil, domain.ErrInvalidAvatar
	}

	normalizedContentType := normalizeAvatarContentType(contentType)
	if normalizedContentType == "" {
		normalizedContentType = normalizeAvatarContentType(http.DetectContentType(avatarBytes))
	}

	ext, ok := avatarExtensionByContentType(normalizedContentType)
	if !ok {
		log.Printf("user.update_profile: unsupported avatar content type user_id=%d original=%q normalized=%q", userID, contentType, normalizedContentType)
		return nil, storagepkg.ErrInvalidFileType
	}

	avatarKey := fmt.Sprintf("users/%d/avatar/%d%s", userID, time.Now().UnixNano(), ext)
	if err := u.avatarStore.PutObject(
		ctx,
		avatarKey,
		bytes.NewReader(avatarBytes),
		int64(len(avatarBytes)),
		normalizedContentType,
	); err != nil {
		return nil, fmt.Errorf("%w: upload avatar object key=%q: %v", domain.ErrInternal, avatarKey, err)
	}

	updatedUser, err := u.userRepo.UpdateAvatarFileKey(ctx, userID, &avatarKey)
	if err != nil {
		return nil, fmt.Errorf("%w: update avatar key in repository key=%q: %v", domain.ErrInternal, avatarKey, err)
	}

	oldAvatarKey := stringValue(user.AvatarFileKey)
	if oldAvatarKey != "" {
		_ = u.avatarStore.DeleteObject(ctx, oldAvatarKey)
	}

	return updatedUser, nil
}

func avatarExtensionByContentType(contentType string) (string, bool) {
	normalizedType := normalizeAvatarContentType(contentType)

	switch normalizedType {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func normalizeAvatarContentType(contentType string) string {
	trimmed := strings.TrimSpace(strings.ToLower(contentType))
	if trimmed == "" {
		return ""
	}

	mediaType, _, err := mime.ParseMediaType(trimmed)
	if err != nil {
		mediaType = trimmed
	}

	if mediaType == "image/jpg" {
		return "image/jpeg"
	}

	return mediaType
}

func (u *AuthUsecase) profileResponse(ctx context.Context, user *domain.User) (domain.ProfileResponse, error) {
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
