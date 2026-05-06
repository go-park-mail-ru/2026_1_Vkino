package routes

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newMultipartFileRequest(t *testing.T, path, field, filename, contentType string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, field, filename))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}

	part, err := writer.CreatePart(header)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestReadSupportFileUploadPayload_Valid(t *testing.T) {
	t.Parallel()

	req := newMultipartFileRequest(t, "/support/files", "file", "note.txt", "text/plain", []byte("hello"))
	rr := httptest.NewRecorder()

	payload, ok := readSupportFileUploadPayload(rr, req)
	require.True(t, ok)
	require.Equal(t, int64(5), payload.SizeBytes)
	require.Equal(t, "note.txt", payload.Filename)
	require.Equal(t, "text/plain", payload.ContentType)
	require.Equal(t, []byte("hello"), payload.Content)
}

func TestReadSupportFileUploadPayload_EmptyFile(t *testing.T) {
	t.Parallel()

	req := newMultipartFileRequest(t, "/support/files", "file", "empty.txt", "text/plain", []byte(""))
	rr := httptest.NewRecorder()

	_, ok := readSupportFileUploadPayload(rr, req)
	require.False(t, ok)
	requireJSONError(t, rr, http.StatusBadRequest, "failed to read support file")
}

func TestReadSupportFileUploadPayload_NotMultipart(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/support/files", bytes.NewReader([]byte("raw")))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()

	_, ok := readSupportFileUploadPayload(rr, req)
	require.False(t, ok)
	requireJSONError(t, rr, http.StatusBadRequest, "invalid multipart form body")
}

func TestSupportFileUploadHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().UploadSupportFile(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, req *supportv1.UploadSupportFileRequest, _ ...any) (*supportv1.UploadSupportFileResponse, error) {
			require.Equal(t, []byte("hello"), req.Content)
			require.Equal(t, "note.txt", req.Filename)
			require.Equal(t, "text/plain", req.ContentType)
			require.Equal(t, int64(5), req.SizeBytes)
			return &supportv1.UploadSupportFileResponse{FileKey: "file-key"}, nil
		})

	handler := newSupportFileUploadHandler(testConfig{}, client)

	req := newMultipartFileRequest(t, "/support/files", "file", "note.txt", "text/plain", []byte("hello"))
	rr := httptest.NewRecorder()

	handler(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestSupportFileURLHandler_InvalidKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newSupportFileURLHandler(testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/files?ticket_id=10", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "file key is required")
}

func TestSupportFileURLHandler_InvalidTicketID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newSupportFileURLHandler(testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/files?key=one&ticket_id=bad", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "ticket id is required")
}

func TestSupportFileURLHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetSupportFileURL(gomock.Any(), &supportv1.GetSupportFileURLRequest{FileKey: "file", TicketId: 12}).
		Return(&supportv1.GetSupportFileURLResponse{FileUrl: "https://cdn/file"}, nil)

	handler := newSupportFileURLHandler(testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/files?key=file&ticket_id=12", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}
