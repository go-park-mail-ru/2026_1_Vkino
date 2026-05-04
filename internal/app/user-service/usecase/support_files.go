//nolint:gocyclo // Validation flow is intentionally explicit for support files.
package usecase

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

const maxSupportFileSize = 10 << 20

func (u *supportUsecase) UploadSupportFile(
	ctx context.Context,
	actorUserID int64,
	req domain2.UploadSupportFileRequest,
) (domain2.SupportFileResponse, error) {
	if u.supportFileStore == nil {
		return domain2.SupportFileResponse{}, storagepkg.ErrStorageUnavailable
	}

	req.ContentType = strings.TrimSpace(req.ContentType)
	req.Filename = strings.TrimSpace(req.Filename)

	if len(req.Content) == 0 || req.SizeBytes <= 0 || int64(len(req.Content)) != req.SizeBytes {
		return domain2.SupportFileResponse{}, domain2.ErrInvalidSupportFilePayload
	}

	if req.SizeBytes > maxSupportFileSize {
		return domain2.SupportFileResponse{}, storagepkg.ErrFileTooLarge
	}

	normalizedContentType := normalizeSupportFileContentType(req.ContentType)
	if normalizedContentType == "" {
		normalizedContentType = normalizeSupportFileContentType(http.DetectContentType(req.Content))
	}

	extension, ok := supportFileExtensionByContentType(normalizedContentType)
	if !ok {
		return domain2.SupportFileResponse{}, storagepkg.ErrInvalidFileType
	}

	fileKey := u.supportFileKey(actorUserID, extension)
	if err := u.supportFileStore.PutObject(
		ctx,
		fileKey,
		bytes.NewReader(req.Content),
		req.SizeBytes,
		normalizedContentType,
	); err != nil {
		return domain2.SupportFileResponse{}, fmt.Errorf(
			"%w: upload support file key=%q: %w",
			domain2.ErrInternal,
			fileKey,
			err,
		)
	}

	fileURL, err := u.supportFileStore.PresignGetObject(ctx, fileKey, 0)
	if err != nil {
		return domain2.SupportFileResponse{}, fmt.Errorf(
			"%w: presign support file key=%q: %w",
			domain2.ErrInternal,
			fileKey,
			err,
		)
	}

	return domain2.SupportFileResponse{
		FileKey:     fileKey,
		FileURL:     fileURL,
		ContentType: normalizedContentType,
		SizeBytes:   req.SizeBytes,
	}, nil
}

func (u *supportUsecase) GetSupportFileURL(
	ctx context.Context,
	actorUserID int64,
	req domain2.GetSupportFileURLRequest,
) (domain2.SupportFileResponse, error) {
	if u.supportFileStore == nil {
		return domain2.SupportFileResponse{}, storagepkg.ErrStorageUnavailable
	}

	req.FileKey = strings.TrimSpace(req.FileKey)
	if req.TicketID <= 0 || req.FileKey == "" {
		return domain2.SupportFileResponse{}, domain2.ErrInvalidSupportFilePayload
	}

	if actorUserID <= 0 {
		return domain2.SupportFileResponse{}, domain2.ErrInvalidToken
	}

	if !strings.HasPrefix(req.FileKey, "support/") {
		return domain2.SupportFileResponse{}, domain2.ErrInvalidSupportFilePayload
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return domain2.SupportFileResponse{}, err
	}

	hasFile, err := u.supportRepo.HasTicketFile(ctx, req.TicketID, req.FileKey)
	if err != nil {
		return domain2.SupportFileResponse{}, fmt.Errorf("%w: verify support file key=%q ticket_id=%d: %w",
			domain2.ErrInternal, req.FileKey, req.TicketID, err)
	}

	if !hasFile {
		return domain2.SupportFileResponse{}, domain2.ErrAccessDenied
	}

	fileURL, err := u.supportFileStore.PresignGetObject(ctx, req.FileKey, 0)
	if err != nil {
		return domain2.SupportFileResponse{}, fmt.Errorf(
			"%w: presign support file key=%q: %w",
			domain2.ErrInternal,
			req.FileKey,
			err,
		)
	}

	return domain2.SupportFileResponse{
		FileKey: req.FileKey,
		FileURL: fileURL,
	}, nil
}

func supportFileExtensionByContentType(contentType string) (string, bool) {
	switch normalizeSupportFileContentType(contentType) {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/webp":
		return ".webp", true
	case "application/pdf":
		return ".pdf", true
	default:
		return "", false
	}
}

func normalizeSupportFileContentType(contentType string) string {
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

func (u *supportUsecase) supportFileKey(actorUserID int64, extension string) string {
	now := time.Now().UnixNano()
	if u.clockService != nil {
		now = u.clockService.Now().UnixNano()
	}

	if actorUserID > 0 {
		return fmt.Sprintf("support/users/%d/%d%s", actorUserID, now, extension)
	}

	return fmt.Sprintf("support/guest/%d%s", now, extension)
}
