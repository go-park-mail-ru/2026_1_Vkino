package routes

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/textproto"
)

func TestIsAvatarReferencePayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        []byte
		contentType string
		want        bool
	}{
		{
			name:        "blob url reference",
			body:        []byte("blob:http://localhost:3000/9929373c-4105-test"),
			contentType: "",
			want:        true,
		},
		{
			name:        "https url reference",
			body:        []byte("https://cdn.example/avatar.jpg"),
			contentType: "text/plain",
			want:        true,
		},
		{
			name:        "empty payload",
			body:        []byte("  "),
			contentType: "",
			want:        true,
		},
		{
			name:        "null payload with image content type",
			body:        []byte("null"),
			contentType: "image/png",
			want:        true,
		},
		{
			name:        "real image payload",
			body:        []byte{0x89, 0x50, 0x4e, 0x47},
			contentType: "image/png",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isAvatarReferencePayload(tt.body, tt.contentType)
			if got != tt.want {
				t.Fatalf("isAvatarReferencePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadUpdateProfilePayload_IgnoresNullAvatarField(t *testing.T) {
	t.Parallel()

	req := newUpdateProfileMultipartRequest(t)

	rr := httptest.NewRecorder()

	payload, ok := readUpdateProfilePayload(rr, req)
	if !ok {
		t.Fatal("readUpdateProfilePayload returned ok=false")
	}

	if payload.Birthdate != "2004-03-01" {
		t.Fatalf("birthdate = %q, want %q", payload.Birthdate, "2004-03-01")
	}

	if len(payload.Avatar) != 0 {
		t.Fatalf("avatar length = %d, want 0", len(payload.Avatar))
	}

	if payload.AvatarContentType != "" {
		t.Fatalf("avatar content-type = %q, want empty", payload.AvatarContentType)
	}
}

func newUpdateProfileMultipartRequest(t *testing.T) *http.Request {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	writeUpdateProfileField(t, writer, "birthdate", "2004-03-01")
	writeUpdateProfileField(t, writer, "avatar", "null")

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/user/profile", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func writeUpdateProfileField(t *testing.T, writer *multipart.Writer, field, value string) {
	t.Helper()

	if err := writer.WriteField(field, value); err != nil {
		t.Fatalf("WriteField %s: %v", field, err)
	}
}

func TestShouldIgnoreMultipartAvatarHeader(t *testing.T) {
	t.Parallel()

	header := &multipart.FileHeader{
		Filename: "",
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"application/octet-stream"},
		},
	}

	if !shouldIgnoreMultipartAvatarHeader(header) {
		t.Fatal("expected multipart avatar header to be ignored")
	}
}
