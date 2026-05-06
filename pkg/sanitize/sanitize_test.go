package sanitize

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestNewAvatarObjectKey(t *testing.T) {
	t.Parallel()

	if _, err := NewAvatarObjectKey(1, ""); err == nil {
		t.Fatal("expected error for empty extension")
	}

	key, err := NewAvatarObjectKey(42, ".png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key == "" {
		t.Fatal("expected key")
	}
}

func TestNormalizeAvatarContentType(t *testing.T) {
	t.Parallel()

	if NormalizeAvatarContentType("image/jpg") != "image/jpeg" {
		t.Fatal("expected jpg to normalize")
	}
}

func TestSanitizeAvatarUploadPNG(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	data := buf.Bytes()
	cleaned, contentType, ext, err := SanitizeAvatarUpload(data, "image/png")
	if err != nil {
		t.Fatalf("sanitize error: %v", err)
	}

	if len(cleaned) == 0 {
		t.Fatal("expected sanitized bytes")
	}

	if contentType != "image/png" || ext != ".png" {
		t.Fatalf("unexpected contentType/ext: %s %s", contentType, ext)
	}
}
