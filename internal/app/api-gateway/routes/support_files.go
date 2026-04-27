package routes

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

const maxSupportFileSize = 10 << 20

type supportFileResponse struct {
	FileKey     string `json:"file_key"`
	FileURL     string `json:"file_url"`
	ContentType string `json:"content_type,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
}

func newSupportFileUploadHandler(fileStore storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fileStore == nil {
			httppkg.ErrResponse(w, http.StatusServiceUnavailable, "support file storage is not configured")

			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxSupportFileSize)
		if err := r.ParseMultipartForm(maxSupportFileSize); err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid multipart form body")

			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid support file")

			return
		}
		defer func() {
			_ = file.Close()
		}()

		payload, err := io.ReadAll(file)
		if err != nil || len(payload) == 0 {
			httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read support file")

			return
		}

		contentType := normalizeSupportFileContentType("")
		if header != nil {
			contentType = normalizeSupportFileContentType(header.Header.Get("Content-Type"))
		}
		if contentType == "" {
			contentType = normalizeSupportFileContentType(http.DetectContentType(payload))
		}

		extension, ok := supportFileExtensionByContentType(contentType)
		if !ok {
			httppkg.ErrResponse(w, http.StatusBadRequest, "unsupported support file type")

			return
		}

		fileKey := fmt.Sprintf("support/%d%s", time.Now().UnixNano(), extension)
		if err = fileStore.PutObject(
			r.Context(),
			fileKey,
			bytes.NewReader(payload),
			int64(len(payload)),
			contentType,
		); err != nil {
			logger.FromContext(r.Context()).
				WithField("file_key", fileKey).
				WithField("content_type", contentType).
				WithField("error", err).
				Error("failed to upload support file")

			httppkg.ErrResponse(w, http.StatusInternalServerError, "failed to upload support file")

			return
		}

		fileURL, err := fileStore.PresignGetObject(r.Context(), fileKey, 0)
		if err != nil {
			logger.FromContext(r.Context()).
				WithField("file_key", fileKey).
				WithField("error", err).
				Error("failed to presign support file")

			httppkg.ErrResponse(w, http.StatusInternalServerError, "failed to create support file url")

			return
		}

		httppkg.Response(w, http.StatusCreated, supportFileResponse{
			FileKey:     fileKey,
			FileURL:     fileURL,
			ContentType: contentType,
			SizeBytes:   int64(len(payload)),
		})
	}
}

func newSupportFileURLHandler(fileStore storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fileStore == nil {
			httppkg.ErrResponse(w, http.StatusServiceUnavailable, "support file storage is not configured")

			return
		}

		fileKey := strings.TrimSpace(r.URL.Query().Get("key"))
		if fileKey == "" {
			httppkg.ErrResponse(w, http.StatusBadRequest, "file key is required")

			return
		}

		fileURL, err := fileStore.PresignGetObject(r.Context(), fileKey, 0)
		if err != nil {
			logger.FromContext(r.Context()).
				WithField("file_key", fileKey).
				WithField("error", err).
				Error("failed to presign support file")

			httppkg.ErrResponse(w, http.StatusInternalServerError, "failed to create support file url")

			return
		}

		httppkg.Response(w, http.StatusOK, supportFileResponse{
			FileKey: fileKey,
			FileURL: fileURL,
		})
	}
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
