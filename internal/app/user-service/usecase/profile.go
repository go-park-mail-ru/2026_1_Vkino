package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/sanitize"
)

const profileCoinsHistoryLimit int32 = 20

func (u *UserUsecase) GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, domain.ErrInvalidToken
	}

	if _, err = u.userRepo.GrantDailyVKinoCoins(ctx, userID); err != nil {
		return domain.ProfileResponse{}, fmt.Errorf("%w: grant daily vkino coins: %w", domain.ErrInternal, err)
	}

	balance, err := u.userRepo.GetVKinoCoinsBalance(ctx, userID)
	if err != nil {
		return domain.ProfileResponse{}, fmt.Errorf("%w: get vkino coins balance: %w", domain.ErrInternal, err)
	}

	history, totalCount, err := u.userRepo.GetVKinoCoinsHistory(ctx, userID, profileCoinsHistoryLimit, 0)
	if err != nil {
		return domain.ProfileResponse{}, fmt.Errorf("%w: get vkino coins history: %w", domain.ErrInternal, err)
	}

	return u.profileResponse(ctx, user, balance, history, totalCount)
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

	return u.profileResponse(ctx, user, 0, nil, 0)
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

	if err := handleAvatarUpdatePreconditions(body, size, u.avatarStore); err != nil {
		if shouldSkipAvatarUpdate(err) {
			return user, nil
		}

		return nil, err
	}

	avatarKey, err := u.resolveAvatarKey(ctx, userID, requestLogger, body, contentType)
	if err != nil {
		if shouldSkipAvatarUpdate(err) {
			return user, nil
		}

		return nil, err
	}

	updatedUser, err := u.persistAvatarUpdate(ctx, userID, user, avatarKey, requestLogger)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

var (
	errSkipAvatarUpdate    = errors.New("skip avatar update")
	errIgnoreAvatarPayload = errors.New("ignore avatar payload")
)

func handleAvatarUpdatePreconditions(body io.Reader, size int64, avatarStore any) error {
	if body == nil || size <= 0 {
		return errSkipAvatarUpdate
	}

	if avatarStore == nil {
		return fmt.Errorf("%w: avatar storage is not configured", domain.ErrInternal)
	}

	return nil
}

func validateDetectedAvatarType(log *logger.Logger, contentType string, avatarBytes []byte) (string, bool) {
	requestedContentType := sanitize.NormalizeAvatarContentType(contentType)
	detectedContentType := sanitize.DetectAvatarContentType(avatarBytes)

	if _, ok := sanitize.AvatarExtensionByContentType(detectedContentType); ok {
		return requestedContentType, true
	}

	log.
		WithField("avatar_content_type", contentType).
		WithField("detected_content_type", detectedContentType).
		WithField("avatar_size", len(avatarBytes)).
		Warn("ignoring unsupported avatar payload during profile update")

	return "", false
}

func shouldSkipAvatarUpdate(err error) bool {
	return errors.Is(err, errSkipAvatarUpdate) || errors.Is(err, errIgnoreAvatarPayload)
}

func (u *UserUsecase) resolveAvatarKey(
	ctx context.Context,
	userID int64,
	log *logger.Logger,
	body io.Reader,
	contentType string,
) (string, error) {
	avatarBytes, err := readAvatarPayload(body, log, contentType)
	if err != nil {
		return "", err
	}

	return u.processAvatarPayload(ctx, userID, log, avatarBytes, contentType)
}

func sanitizeAvatarPayload(
	log *logger.Logger,
	contentType string,
	avatarBytes []byte,
) ([]byte, string, string, error) {
	detectedContentType := sanitize.DetectAvatarContentType(avatarBytes)
	requestedContentType, ok := validateDetectedAvatarType(log, contentType, avatarBytes)

	if !ok {
		return nil, "", "", errIgnoreAvatarPayload
	}

	sanitizedAvatarBytes, normalizedContentType, ext, err := sanitize.SanitizeAvatarUpload(
		avatarBytes,
		requestedContentType,
	)
	if err != nil {
		log.
			WithField("original_content_type", contentType).
			WithField("requested_content_type", requestedContentType).
			WithField("detected_content_type", detectedContentType).
			WithField("avatar_size", len(avatarBytes)).
			WithField("avatar_preview", string(bytes.TrimSpace(avatarBytes))).
			WithField("error", err).
			Error("invalid avatar payload")

		return nil, "", "", err
	}

	return sanitizedAvatarBytes, normalizedContentType, ext, nil
}

func (u *UserUsecase) persistAvatarUpdate(
	ctx context.Context,
	userID int64,
	user *domain.User,
	avatarKey string,
	log *logger.Logger,
) (*domain.User, error) {
	updatedUser, err := u.userRepo.UpdateAvatarFileKey(ctx, userID, &avatarKey)
	if err != nil {
		return nil, fmt.Errorf("%w: update avatar key in repository key=%q: %w", domain.ErrInternal, avatarKey, err)
	}

	oldAvatarKey := stringValue(user.AvatarFileKey)
	if oldAvatarKey == "" {
		return updatedUser, nil
	}

	if err = u.avatarStore.DeleteObject(ctx, oldAvatarKey); err != nil {
		log.
			WithField("avatar_key", oldAvatarKey).
			WithField("error", err).
			Warn("failed to delete previous avatar")
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

func readAvatarPayload(body io.Reader, log *logger.Logger, contentType string) ([]byte, error) {
	avatarBytes, err := io.ReadAll(body)
	if shouldIgnoreEmptyAvatar(err, avatarBytes, log) {
		return nil, errIgnoreAvatarPayload
	}

	if ignore := shouldIgnoreAvatarPayload(avatarBytes, contentType); ignore {
		logIgnoredAvatarPayload(log, contentType, avatarBytes)

		return nil, errIgnoreAvatarPayload
	}

	return avatarBytes, nil
}

func (u *UserUsecase) processAvatarPayload(
	ctx context.Context,
	userID int64,
	log *logger.Logger,
	avatarBytes []byte,
	contentType string,
) (string, error) {
	sanitizedAvatarBytes, normalizedContentType, ext, err := sanitizeAvatarPayload(
		log,
		contentType,
		avatarBytes,
	)
	if err != nil {
		return "", err
	}

	return u.storeAvatar(ctx, userID, sanitizedAvatarBytes, normalizedContentType, ext)
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

func shouldIgnoreEmptyAvatar(err error, avatarBytes []byte, log *logger.Logger) bool {
	if err == nil && len(avatarBytes) > 0 {
		return false
	}

	if err != nil {
		log.WithField("error", err).Error("failed to read avatar body")
	}

	return true
}

func logIgnoredAvatarPayload(log *logger.Logger, contentType string, avatarBytes []byte) {
	log.
		WithField("avatar_content_type", contentType).
		WithField("avatar_size", len(avatarBytes)).
		WithField("avatar_preview", string(bytes.TrimSpace(avatarBytes))).
		Info("ignoring avatar payload during profile update")
}

func (u *UserUsecase) storeAvatar(
	ctx context.Context,
	userID int64,
	sanitizedAvatarBytes []byte,
	normalizedContentType, ext string,
) (string, error) {
	avatarKey, err := sanitize.NewAvatarObjectKey(userID, ext)
	if err != nil {
		return "", fmt.Errorf("%w: generate avatar object key: %w", domain.ErrInternal, err)
	}

	if err := u.avatarStore.PutObject(
		ctx,
		avatarKey,
		bytes.NewReader(sanitizedAvatarBytes),
		int64(len(sanitizedAvatarBytes)),
		normalizedContentType,
	); err != nil {
		return "", fmt.Errorf("%w: upload avatar object key=%q: %w", domain.ErrInternal, avatarKey, err)
	}

	return avatarKey, nil
}
