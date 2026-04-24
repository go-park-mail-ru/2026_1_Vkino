package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func (u *UserUsecase) GetProfile(ctx context.Context, userID int64) (domain2.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain2.ProfileResponse{}, domain2.ErrInvalidToken
	}

	return u.profileResponse(ctx, user)
}

func (u *UserUsecase) UpdateProfile(
	ctx context.Context,
	userID int64,
	birthdate string,
	body io.Reader,
	size int64,
	contentType string,
) (domain2.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain2.ProfileResponse{}, domain2.ErrInvalidToken
	}

	user, err = u.updateBirthdateIfProvided(ctx, userID, birthdate, user)
	if err != nil {
		return domain2.ProfileResponse{}, err
	}

	user, err = u.updateAvatarIfProvided(ctx, userID, user, body, size, contentType)
	if err != nil {
		return domain2.ProfileResponse{}, err
	}

	return u.profileResponse(ctx, user)
}

func (u *UserUsecase) updateBirthdateIfProvided(
	ctx context.Context,
	userID int64,
	birthdate string,
	user *domain2.User,
) (*domain2.User, error) {
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
		return nil, fmt.Errorf("%w: update birthdate in repository: %v", domain2.ErrInternal, err)
	}

	return updatedUser, nil
}

func parseBirthdate(rawBirthdate string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02", rawBirthdate)
	if err != nil || parsed.After(time.Now()) {
		return nil, domain2.ErrInvalidBirthdate
	}

	return &parsed, nil
}

func (u *UserUsecase) updateAvatarIfProvided(
	ctx context.Context,
	userID int64,
	user *domain2.User,
	body io.Reader,
	size int64,
	contentType string,
) (*domain2.User, error) {
	requestLogger := logger.FromContext(ctx).
		WithField("usecase", "UserUsecase.UpdateProfile")

	if body == nil {
		return user, nil
	}

	if u.avatarStore == nil {
		return nil, fmt.Errorf("%w: avatar storage is not configured", domain2.ErrInternal)
	}

	if size <= 0 {
		return nil, domain2.ErrInvalidAvatar
	}

	avatarBytes, err := io.ReadAll(body)
	if err != nil || len(avatarBytes) == 0 {
		if err != nil {
			requestLogger.
				WithField("error", err).
				Error("failed to read avatar body")
		}

		return nil, domain2.ErrInvalidAvatar
	}

	normalizedContentType := normalizeAvatarContentType(contentType)
	if normalizedContentType == "" {
		normalizedContentType = normalizeAvatarContentType(http.DetectContentType(avatarBytes))
	}

	ext, ok := avatarExtensionByContentType(normalizedContentType)
	if !ok {
		requestLogger.
			WithField("original_content_type", contentType).
			WithField("normalized_content_type", normalizedContentType).
			Error("unsupported avatar content type")

		return nil, storagepkg.ErrInvalidFileType
	}

	avatarKey := fmt.Sprintf("users/%d/avatar/%d%s", userID, u.clockService.Now().UnixNano(), ext)
	if err := u.avatarStore.PutObject(
		ctx,
		avatarKey,
		bytes.NewReader(avatarBytes),
		int64(len(avatarBytes)),
		normalizedContentType,
	); err != nil {
		return nil, fmt.Errorf("%w: upload avatar object key=%q: %v", domain2.ErrInternal, avatarKey, err)
	}

	updatedUser, err := u.userRepo.UpdateAvatarFileKey(ctx, userID, &avatarKey)
	if err != nil {
		return nil, fmt.Errorf("%w: update avatar key in repository key=%q: %v", domain2.ErrInternal, avatarKey, err)
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
