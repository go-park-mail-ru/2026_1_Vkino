//nolint:gocyclo,lll // Profile update flow stays explicit for validation clarity.
package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/sanitize"
)

func (u *UserUsecase) GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, domain.ErrInvalidToken
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

func (u *UserUsecase) updateBirthdateIfProvided(
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
		return nil, fmt.Errorf("%w: update birthdate in repository: %w", domain.ErrInternal, err)
	}

	return updatedUser, nil
}

//nolint:cyclop // Avatar update validation intentionally stays explicit.
func (u *UserUsecase) updateAvatarIfProvided(
	ctx context.Context,
	userID int64,
	user *domain.User,
	body io.Reader,
	size int64,
	contentType string,
) (*domain.User, error) {
	requestLogger := logger.FromContext(ctx).
		WithField("usecase", "UserUsecase.UpdateProfile")

	if body == nil {
		return user, nil
	}

	if u.avatarStore == nil {
		return nil, fmt.Errorf("%w: avatar storage is not configured", domain.ErrInternal)
	}

	if size <= 0 {
		return user, nil
	}

	avatarBytes, err := io.ReadAll(body)
	if err != nil || len(avatarBytes) == 0 {
		if err != nil {
			requestLogger.
				WithField("error", err).
				Error("failed to read avatar body")
		}

		return user, nil
	}

	if shouldIgnoreAvatarPayload(avatarBytes, contentType) {
		requestLogger.
			WithField("avatar_content_type", contentType).
			WithField("avatar_size", len(avatarBytes)).
			WithField("avatar_preview", string(bytes.TrimSpace(avatarBytes))).
			Info("ignoring avatar payload during profile update")

		return user, nil
	}

	requestedContentType := sanitize.NormalizeAvatarContentType(contentType)
	detectedContentType := sanitize.DetectAvatarContentType(avatarBytes)
	if _, ok := sanitize.AvatarExtensionByContentType(detectedContentType); !ok {
		requestLogger.
			WithField("avatar_content_type", contentType).
			WithField("detected_content_type", detectedContentType).
			WithField("avatar_size", len(avatarBytes)).
			Warn("ignoring unsupported avatar payload during profile update")

		return user, nil
	}

	sanitizedAvatarBytes, normalizedContentType, ext, err := sanitize.SanitizeAvatarUpload(avatarBytes, requestedContentType)
	if err != nil {
		requestLogger.
			WithField("original_content_type", contentType).
			WithField("requested_content_type", requestedContentType).
			WithField("detected_content_type", detectedContentType).
			WithField("avatar_size", len(avatarBytes)).
			WithField("avatar_preview", string(bytes.TrimSpace(avatarBytes))).
			WithField("error", err).
			Error("invalid avatar payload")

		return nil, err
	}

	avatarKey, err := sanitize.NewAvatarObjectKey(userID, ext)
	if err != nil {
		return nil, fmt.Errorf("%w: generate avatar object key: %w", domain.ErrInternal, err)
	}

	err = u.avatarStore.PutObject(
		ctx,
		avatarKey,
		bytes.NewReader(sanitizedAvatarBytes),
		int64(len(sanitizedAvatarBytes)),
		normalizedContentType,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: upload avatar object key=%q: %w", domain.ErrInternal, avatarKey, err)
	}

	updatedUser, err := u.userRepo.UpdateAvatarFileKey(ctx, userID, &avatarKey)
	if err != nil {
		return nil, fmt.Errorf("%w: update avatar key in repository key=%q: %w", domain.ErrInternal, avatarKey, err)
	}

	oldAvatarKey := stringValue(user.AvatarFileKey)
	if oldAvatarKey != "" {
		if err = u.avatarStore.DeleteObject(ctx, oldAvatarKey); err != nil {
			requestLogger.
				WithField("avatar_key", oldAvatarKey).
				WithField("error", err).
				Warn("failed to delete previous avatar")
		}
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

func shouldIgnoreAvatarPayload(body []byte, contentType string) bool {
	trimmedBody := bytes.TrimSpace(body)
	if len(trimmedBody) == 0 {
		return true
	}

	value := strings.ToLower(string(trimmedBody))

	if value == "null" || value == "undefined" {
		return true
	}

	if strings.HasPrefix(value, "blob:") ||
		strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") {
		return true
	}

	trimmedType := strings.ToLower(strings.TrimSpace(contentType))

	return !strings.HasPrefix(trimmedType, "image/")
}
