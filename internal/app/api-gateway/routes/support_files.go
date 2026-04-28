package routes

import (
	"errors"
	"io"
	"net/http"
	"strings"

	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

const (
	maxSupportFileSize          = 10 << 20
	maxSupportMultipartOverhead = 1 << 20
)

type supportFileUploadPayload struct {
	Content     []byte
	Filename    string
	ContentType string
	SizeBytes   int64
}

func newSupportFileUploadHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, ok := readSupportFileUploadPayload(w, r)
		if !ok {
			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.UploadSupportFile(r.Context(), &supportv1.UploadSupportFileRequest{
			Content:     payload.Content,
			Filename:    payload.Filename,
			ContentType: payload.ContentType,
			SizeBytes:   payload.SizeBytes,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, resp)
	}
}

func newSupportFileURLHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileKey := strings.TrimSpace(r.URL.Query().Get("key"))
		if fileKey == "" {
			httppkg.ErrResponse(w, http.StatusBadRequest, "file key is required")

			return
		}

		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetSupportFileURL(r.Context(), &supportv1.GetSupportFileURLRequest{
			FileKey: fileKey,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, resp)
	}
}

func readSupportFileUploadPayload(w http.ResponseWriter, r *http.Request) (supportFileUploadPayload, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSupportFileSize+maxSupportMultipartOverhead)
	if err := r.ParseMultipartForm(maxSupportFileSize); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httppkg.ErrResponse(w, http.StatusRequestEntityTooLarge, "support file exceeds the size limit")

			return supportFileUploadPayload{}, false
		}

		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid multipart form body")

		return supportFileUploadPayload{}, false
	}

	if r.MultipartForm != nil {
		defer func() {
			_ = r.MultipartForm.RemoveAll()
		}()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid support file")

		return supportFileUploadPayload{}, false
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(io.LimitReader(file, maxSupportFileSize+1))
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read support file")

		return supportFileUploadPayload{}, false
	}

	if int64(len(content)) > maxSupportFileSize {
		httppkg.ErrResponse(w, http.StatusRequestEntityTooLarge, "support file exceeds the size limit")

		return supportFileUploadPayload{}, false
	}

	if len(content) == 0 {
		httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read support file")

		return supportFileUploadPayload{}, false
	}

	payload := supportFileUploadPayload{
		Content:   content,
		SizeBytes: int64(len(content)),
	}

	if header != nil {
		payload.Filename = header.Filename
		payload.ContentType = header.Header.Get("Content-Type")
	}

	return payload, true
}
