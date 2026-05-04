//nolint:gocyclo // Binary format parsing requires explicit branch-heavy validation.
package sanitize

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

const (
	maxAvatarWidth  = 4096
	maxAvatarHeight = 4096
	maxAvatarPixels = maxAvatarWidth * maxAvatarHeight
	jpegQuality     = 90
)

const (
	vp8LDimensionMask  = 0x3fff
	vp8LHeightBitShift = 14
	vp8PayloadMinSize  = 10
	vp8LPayloadMinSize = 5
	vp8xChunkMinSize   = 10
	webpAnimationFlag  = 0x02
	vp8StartCodeByte1  = 0x9d
	vp8StartCodeByte2  = 0x01
	vp8StartCodeByte3  = 0x2a
	vp8LSignature      = 0x2f
	webpChunkHeaderLen = 8
)

const (
	avatarContentTypePNG  = "image/png"
	avatarContentTypeJPEG = "image/jpeg"
	avatarContentTypeJPG  = "image/jpg"
	avatarContentTypeWebP = "image/webp"
)

var (
	errEmptyAvatarExtension   = errors.New("empty avatar extension")
	errInvalidWebPHeader      = errors.New("invalid webp header")
	errInvalidWebPRIFFSize    = errors.New("invalid webp riff size")
	errInvalidWebPChunkHeader = errors.New("invalid webp chunk header")
	errInvalidWebPChunkSize   = errors.New("invalid webp chunk size")
	errInvalidWebPVP8XChunk   = errors.New("invalid webp vp8x chunk")
	errAnimatedWebPNotAllowed = errors.New("animated webp is not allowed")
	errWebPFrameExceedsCanvas = errors.New("webp frame exceeds canvas")
	errUnsupportedWebPChunk   = errors.New("unsupported webp chunk")
	errMissingWebPImageChunk  = errors.New("missing webp image chunk")
	errInvalidVP8Payload      = errors.New("invalid vp8 payload")
	errInvalidVP8StartCode    = errors.New("invalid vp8 start code")
	errInvalidVP8Dimensions   = errors.New("invalid vp8 dimensions")
	errInvalidVP8LPayload     = errors.New("invalid vp8l payload")
	errInvalidVP8LSignature   = errors.New("invalid vp8l signature")
	errInvalidVP8LDimensions  = errors.New("invalid vp8l dimensions")
)

func NewAvatarObjectKey(userID int64, ext string) (string, error) {
	if ext == "" {
		return "", errEmptyAvatarExtension
	}

	var suffix [16]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("read avatar suffix: %w", err)
	}

	return fmt.Sprintf("users/%d/avatar/%s%s", userID, hex.EncodeToString(suffix[:]), ext), nil
}

func SanitizeAvatarUpload(
	avatarBytes []byte,
	contentType string,
) ([]byte, string, string, error) {
	detectedContentType := DetectAvatarContentType(avatarBytes)

	ext, ok := AvatarExtensionByContentType(detectedContentType)
	if !ok {
		return nil, "", "", ErrInvalidFileType
	}

	if contentType != "" && contentType != detectedContentType {
		return nil, "", "", ErrInvalidFileType
	}

	sanitizedAvatarBytes, err := sanitizeAvatarBytes(avatarBytes, detectedContentType)
	if err != nil {
		return nil, "", "", err
	}

	return sanitizedAvatarBytes, detectedContentType, ext, nil
}

func DetectAvatarContentType(avatarBytes []byte) string {
	detectedContentType := NormalizeAvatarContentType(http.DetectContentType(avatarBytes))
	if _, ok := AvatarExtensionByContentType(detectedContentType); ok {
		return detectedContentType
	}

	if hasWebPHeader(avatarBytes) {
		return avatarContentTypeWebP
	}

	return detectedContentType
}

func NormalizeAvatarContentType(contentType string) string {
	trimmed := strings.TrimSpace(strings.ToLower(contentType))
	if trimmed == "" {
		return ""
	}

	mediaType, _, err := mime.ParseMediaType(trimmed)
	if err != nil {
		mediaType = trimmed
	}

	if mediaType == avatarContentTypeJPG {
		return avatarContentTypeJPEG
	}

	return mediaType
}

func AvatarExtensionByContentType(contentType string) (string, bool) {
	normalizedType := NormalizeAvatarContentType(contentType)

	switch normalizedType {
	case avatarContentTypePNG:
		return ".png", true
	case avatarContentTypeJPEG:
		return ".jpg", true
	case avatarContentTypeWebP:
		return ".webp", true
	default:
		return "", false
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
		return nil, fmt.Errorf("%w: encode avatar image: %w", domain.ErrInternal, err)
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

//nolint:gocognit,gocyclo,cyclop // WebP container parsing needs explicit chunk-by-chunk validation.
func webpDimensions(avatarBytes []byte) (int, int, error) {
	if !hasWebPHeader(avatarBytes) || len(avatarBytes) < 20 {
		return 0, 0, errInvalidWebPHeader
	}

	riffSize := int(binary.LittleEndian.Uint32(avatarBytes[4:8]))
	if riffSize+8 != len(avatarBytes) {
		return 0, 0, errInvalidWebPRIFFSize
	}

	var (
		canvasWidth  int
		canvasHeight int
	)

	for offset := 12; offset < len(avatarBytes); {
		if offset+webpChunkHeaderLen > len(avatarBytes) {
			return 0, 0, errInvalidWebPChunkHeader
		}

		chunkType := string(avatarBytes[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(avatarBytes[offset+4 : offset+8]))
		offset += webpChunkHeaderLen

		if chunkSize < 0 || offset+chunkSize > len(avatarBytes) {
			return 0, 0, errInvalidWebPChunkSize
		}

		chunkPayload := avatarBytes[offset : offset+chunkSize]

		switch chunkType {
		case "VP8X":
			if chunkSize < vp8xChunkMinSize {
				return 0, 0, errInvalidWebPVP8XChunk
			}

			if chunkPayload[0]&webpAnimationFlag != 0 {
				return 0, 0, errAnimatedWebPNotAllowed
			}

			canvasWidth = 1 + int(uint32(chunkPayload[4])|uint32(chunkPayload[5])<<8|uint32(chunkPayload[6])<<16)
			canvasHeight = 1 + int(uint32(chunkPayload[7])|uint32(chunkPayload[8])<<8|uint32(chunkPayload[9])<<16)
		case "VP8 ":
			width, height, err := vp8Dimensions(chunkPayload)
			if err != nil {
				return 0, 0, err
			}

			if canvasWidth > 0 && (width > canvasWidth || height > canvasHeight) {
				return 0, 0, errWebPFrameExceedsCanvas
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
				return 0, 0, errWebPFrameExceedsCanvas
			}

			if canvasWidth > 0 && canvasHeight > 0 {
				return canvasWidth, canvasHeight, nil
			}

			return width, height, nil
		case "ALPH":
		default:
			return 0, 0, fmt.Errorf("%w: %q", errUnsupportedWebPChunk, chunkType)
		}

		offset += chunkSize
		if chunkSize%2 != 0 {
			offset++
		}
	}

	return 0, 0, errMissingWebPImageChunk
}

func hasWebPHeader(avatarBytes []byte) bool {
	return len(avatarBytes) >= 12 &&
		string(avatarBytes[:4]) == "RIFF" &&
		string(avatarBytes[8:12]) == "WEBP"
}

func vp8Dimensions(chunkPayload []byte) (int, int, error) {
	if len(chunkPayload) < vp8PayloadMinSize {
		return 0, 0, errInvalidVP8Payload
	}

	if chunkPayload[3] != vp8StartCodeByte1 ||
		chunkPayload[4] != vp8StartCodeByte2 ||
		chunkPayload[5] != vp8StartCodeByte3 {
		return 0, 0, errInvalidVP8StartCode
	}

	width := int(binary.LittleEndian.Uint16(chunkPayload[6:8]) & vp8LDimensionMask)

	height := int(binary.LittleEndian.Uint16(chunkPayload[8:10]) & vp8LDimensionMask)
	if width <= 0 || height <= 0 {
		return 0, 0, errInvalidVP8Dimensions
	}

	return width, height, nil
}

func vp8LDimensions(chunkPayload []byte) (int, int, error) {
	if len(chunkPayload) < vp8LPayloadMinSize {
		return 0, 0, errInvalidVP8LPayload
	}

	if chunkPayload[0] != vp8LSignature {
		return 0, 0, errInvalidVP8LSignature
	}

	bits := binary.LittleEndian.Uint32(chunkPayload[1:5])
	rawWidth := bits & vp8LDimensionMask
	rawHeight := (bits >> vp8LHeightBitShift) & vp8LDimensionMask
	width := 1 + int(rawWidth)

	height := 1 + int(rawHeight)
	if width <= 0 || height <= 0 {
		return 0, 0, errInvalidVP8LDimensions
	}

	return width, height, nil
}

// PNG and JPEG are re-encoded to strip user-controlled metadata and trailing data.
func sanitizeAvatarBytes(avatarBytes []byte, contentType string) ([]byte, error) {
	switch contentType {
	case avatarContentTypePNG:
		return sanitizeDecodedAvatar(
			avatarBytes,
			png.DecodeConfig,
			png.Decode,
			func(w io.Writer, img image.Image) error {
				encoder := png.Encoder{CompressionLevel: png.DefaultCompression}

				return encoder.Encode(w, img)
			},
		)
	case avatarContentTypeJPEG:
		return sanitizeDecodedAvatar(
			avatarBytes,
			jpeg.DecodeConfig,
			jpeg.Decode,
			func(w io.Writer, img image.Image) error {
				return jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality})
			},
		)
	case avatarContentTypeWebP:
		if err := validateWebPAvatar(avatarBytes); err != nil {
			return nil, err
		}

		return avatarBytes, nil
	default:
		return nil, ErrInvalidFileType
	}
}
