package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

const (
	maxAvatarWidth  = 4096
	maxAvatarHeight = 4096
	maxAvatarPixels = maxAvatarWidth * maxAvatarHeight
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
	requestLogger := logger.FromContext(ctx).
		WithField("usecase", "AuthUsecase.UpdateProfile")

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
			requestLogger.
				WithField("error", err).
				Error("failed to read avatar body")
		}
		return nil, domain.ErrInvalidAvatar
	}

	requestedContentType := normalizeAvatarContentType(contentType)
	detectedContentType := detectAvatarContentType(avatarBytes)
	sanitizedAvatarBytes, normalizedContentType, ext, err := sanitizeAvatarUpload(avatarBytes, requestedContentType)
	if err != nil {
		requestLogger.
			WithField("original_content_type", contentType).
			WithField("requested_content_type", requestedContentType).
			WithField("detected_content_type", detectedContentType).
			WithField("error", err).
			Error("invalid avatar payload")
		return nil, err
	}

	avatarKey, err := newAvatarObjectKey(userID, ext)
	if err != nil {
		return nil, fmt.Errorf("%w: generate avatar object key: %v", domain.ErrInternal, err)
	}

	if err := u.avatarStore.PutObject(
		ctx,
		avatarKey,
		bytes.NewReader(sanitizedAvatarBytes),
		int64(len(sanitizedAvatarBytes)),
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

func sanitizeAvatarUpload(
	avatarBytes []byte,
	contentType string,
) ([]byte, string, string, error) {
	detectedContentType := detectAvatarContentType(avatarBytes)
	ext, ok := avatarExtensionByContentType(detectedContentType)
	if !ok {
		return nil, "", "", storagepkg.ErrInvalidFileType
	}

	if contentType != "" && contentType != detectedContentType {
		return nil, "", "", storagepkg.ErrInvalidFileType
	}

	sanitizedAvatarBytes, err := sanitizeAvatarBytes(avatarBytes, detectedContentType)
	if err != nil {
		return nil, "", "", err
	}

	return sanitizedAvatarBytes, detectedContentType, ext, nil
}

func detectAvatarContentType(avatarBytes []byte) string {
	detectedContentType := normalizeAvatarContentType(http.DetectContentType(avatarBytes))
	if _, ok := avatarExtensionByContentType(detectedContentType); ok {
		return detectedContentType
	}

	if hasWebPHeader(avatarBytes) {
		return "image/webp"
	}

	return detectedContentType
}

// PNG and JPEG are re-encoded to strip user-controlled metadata and trailing data.
func sanitizeAvatarBytes(avatarBytes []byte, contentType string) ([]byte, error) {
	switch contentType {
	case "image/png":
		return sanitizeDecodedAvatar(
			avatarBytes,
			png.DecodeConfig,
			png.Decode,
			func(w io.Writer, img image.Image) error {
				encoder := png.Encoder{CompressionLevel: png.DefaultCompression}

				return encoder.Encode(w, img)
			},
		)
	case "image/jpeg":
		return sanitizeDecodedAvatar(
			avatarBytes,
			jpeg.DecodeConfig,
			jpeg.Decode,
			func(w io.Writer, img image.Image) error {
				return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
			},
		)
	case "image/webp":
		if err := validateWebPAvatar(avatarBytes); err != nil {
			return nil, err
		}

		return avatarBytes, nil
	default:
		return nil, storagepkg.ErrInvalidFileType
	}
}

func sanitizeDecodedAvatar(
	avatarBytes []byte,
	decodeConfig func(io.Reader) (image.Config, error),
	decode func(io.Reader) (image.Image, error),
	encode func(io.Writer, image.Image) error,
) ([]byte, error) {
	config, err := decodeConfig(bytes.NewReader(avatarBytes))
	if err != nil {
		return nil, domain.ErrInvalidAvatar
	}

	if err = validateAvatarDimensions(config.Width, config.Height); err != nil {
		return nil, err
	}

	img, err := decode(bytes.NewReader(avatarBytes))
	if err != nil {
		return nil, domain.ErrInvalidAvatar
	}

	var buf bytes.Buffer
	if err = encode(&buf, img); err != nil {
		return nil, fmt.Errorf("%w: encode avatar image: %v", domain.ErrInternal, err)
	}

	return buf.Bytes(), nil
}

func validateAvatarDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return domain.ErrInvalidAvatar
	}

	if width > maxAvatarWidth || height > maxAvatarHeight {
		return domain.ErrInvalidAvatar
	}

	if width*height > maxAvatarPixels {
		return domain.ErrInvalidAvatar
	}

	return nil
}

func validateWebPAvatar(avatarBytes []byte) error {
	width, height, err := webpDimensions(avatarBytes)
	if err != nil {
		return domain.ErrInvalidAvatar
	}

	return validateAvatarDimensions(width, height)
}

func webpDimensions(avatarBytes []byte) (int, int, error) {
	if !hasWebPHeader(avatarBytes) || len(avatarBytes) < 20 {
		return 0, 0, fmt.Errorf("invalid webp header")
	}

	riffSize := int(binary.LittleEndian.Uint32(avatarBytes[4:8]))
	if riffSize+8 != len(avatarBytes) {
		return 0, 0, fmt.Errorf("invalid webp riff size")
	}

	var (
		canvasWidth  int
		canvasHeight int
	)

	for offset := 12; offset < len(avatarBytes); {
		if offset+8 > len(avatarBytes) {
			return 0, 0, fmt.Errorf("invalid webp chunk header")
		}

		chunkType := string(avatarBytes[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(avatarBytes[offset+4 : offset+8]))
		offset += 8

		if chunkSize < 0 || offset+chunkSize > len(avatarBytes) {
			return 0, 0, fmt.Errorf("invalid webp chunk size")
		}

		chunkPayload := avatarBytes[offset : offset+chunkSize]

		switch chunkType {
		case "VP8X":
			if chunkSize < 10 {
				return 0, 0, fmt.Errorf("invalid webp vp8x chunk")
			}

			if chunkPayload[0]&0x02 != 0 {
				return 0, 0, fmt.Errorf("animated webp is not allowed")
			}

			canvasWidth = 1 + int(uint32(chunkPayload[4])|uint32(chunkPayload[5])<<8|uint32(chunkPayload[6])<<16)
			canvasHeight = 1 + int(uint32(chunkPayload[7])|uint32(chunkPayload[8])<<8|uint32(chunkPayload[9])<<16)
		case "VP8 ":
			width, height, err := vp8Dimensions(chunkPayload)
			if err != nil {
				return 0, 0, err
			}

			if canvasWidth > 0 && (width > canvasWidth || height > canvasHeight) {
				return 0, 0, fmt.Errorf("webp frame exceeds canvas")
			}

			if canvasWidth > 0 && canvasHeight > 0 {
				return canvasWidth, canvasHeight, nil
			}

			return width, height, nil
		case "VP8L":
			width, height, err := vp8LDimensions(chunkPayload)
			if err != nil {
				return 0, 0, err
			}

			if canvasWidth > 0 && (width > canvasWidth || height > canvasHeight) {
				return 0, 0, fmt.Errorf("webp frame exceeds canvas")
			}

			if canvasWidth > 0 && canvasHeight > 0 {
				return canvasWidth, canvasHeight, nil
			}

			return width, height, nil
		case "ALPH":
		default:
			return 0, 0, fmt.Errorf("unsupported webp chunk %q", chunkType)
		}

		offset += chunkSize
		if chunkSize%2 != 0 {
			offset++
		}
	}

	return 0, 0, fmt.Errorf("missing webp image chunk")
}

func hasWebPHeader(avatarBytes []byte) bool {
	return len(avatarBytes) >= 12 &&
		string(avatarBytes[:4]) == "RIFF" &&
		string(avatarBytes[8:12]) == "WEBP"
}

func vp8Dimensions(chunkPayload []byte) (int, int, error) {
	if len(chunkPayload) < 10 {
		return 0, 0, fmt.Errorf("invalid vp8 payload")
	}

	if chunkPayload[3] != 0x9d || chunkPayload[4] != 0x01 || chunkPayload[5] != 0x2a {
		return 0, 0, fmt.Errorf("invalid vp8 start code")
	}

	width := int(binary.LittleEndian.Uint16(chunkPayload[6:8]) & 0x3fff)
	height := int(binary.LittleEndian.Uint16(chunkPayload[8:10]) & 0x3fff)
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid vp8 dimensions")
	}

	return width, height, nil
}

func vp8LDimensions(chunkPayload []byte) (int, int, error) {
	if len(chunkPayload) < 5 {
		return 0, 0, fmt.Errorf("invalid vp8l payload")
	}

	if chunkPayload[0] != 0x2f {
		return 0, 0, fmt.Errorf("invalid vp8l signature")
	}

	bits := binary.LittleEndian.Uint32(chunkPayload[1:5])
	width := 1 + int(bits&0x3fff)
	height := 1 + int((bits>>14)&0x3fff)
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid vp8l dimensions")
	}

	return width, height, nil
}

func newAvatarObjectKey(userID int64, ext string) (string, error) {
	if ext == "" {
		return "", fmt.Errorf("empty avatar extension")
	}

	var suffix [16]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("read avatar suffix: %w", err)
	}

	return fmt.Sprintf("users/%d/avatar/%s%s", userID, hex.EncodeToString(suffix[:]), ext), nil
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
