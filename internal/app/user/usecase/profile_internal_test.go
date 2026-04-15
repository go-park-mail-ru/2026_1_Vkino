package usecase

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	storagepkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func TestParseBirthdate(t *testing.T) {
	t.Parallel()

	if _, err := parseBirthdate("not-a-date"); err == nil {
		t.Fatal("expected parse error")
	}

	future := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	if _, err := parseBirthdate(future); err == nil {
		t.Fatal("expected future birthdate error")
	}

	got, err := parseBirthdate("2001-09-12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || got.Format("2006-01-02") != "2001-09-12" {
		t.Fatalf("unexpected birthdate: %v", got)
	}
}

func TestAvatarExtensionByContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		contentType string
		wantExt     string
		wantOK      bool
	}{
		{contentType: "image/png", wantExt: ".png", wantOK: true},
		{contentType: "image/jpeg", wantExt: ".jpg", wantOK: true},
		{contentType: "image/webp", wantExt: ".webp", wantOK: true},
		{contentType: "text/plain", wantExt: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			t.Parallel()

			got, ok := avatarExtensionByContentType(tt.contentType)
			if got != tt.wantExt || ok != tt.wantOK {
				t.Fatalf("expected (%q, %v), got (%q, %v)", tt.wantExt, tt.wantOK, got, ok)
			}
		})
	}
}

func TestNormalizeAvatarContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{name: "empty", contentType: "", want: ""},
		{name: "jpg alias", contentType: "image/jpg", want: "image/jpeg"},
		{name: "trimmed with params", contentType: " Image/PNG ; charset=utf-8 ", want: "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeAvatarContentType(tt.contentType); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAuthUsecase_updateBirthdateIfProvided(t *testing.T) {
	t.Parallel()

	u := &AuthUsecase{}
	user := &domain.User{Email: "user@example.com"}

	got, err := u.updateBirthdateIfProvided(context.Background(), 7, "   ", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != user {
		t.Fatal("expected original user to be returned when birthdate is empty")
	}
}

func TestAuthUsecase_updateAvatarIfProvidedValidation(t *testing.T) {
	t.Parallel()

	user := &domain.User{Email: "user@example.com"}

	t.Run("nil body", func(t *testing.T) {
		u := &AuthUsecase{}
		got, err := u.updateAvatarIfProvided(context.Background(), 7, user, nil, 0, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != user {
			t.Fatal("expected original user")
		}
	})

	t.Run("missing store", func(t *testing.T) {
		u := &AuthUsecase{}
		_, err := u.updateAvatarIfProvided(context.Background(), 7, user, strings.NewReader("img"), 3, "image/png")
		if err == nil || !strings.Contains(err.Error(), domain.ErrInternal.Error()) {
			t.Fatalf("expected internal error, got %v", err)
		}
	})

	t.Run("invalid size", func(t *testing.T) {
		u := &AuthUsecase{avatarStore: &storagepkg.S3Storage{}}
		_, err := u.updateAvatarIfProvided(context.Background(), 7, user, strings.NewReader("img"), 0, "image/png")
		if err != domain.ErrInvalidAvatar {
			t.Fatalf("expected ErrInvalidAvatar, got %v", err)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		u := &AuthUsecase{avatarStore: &storagepkg.S3Storage{}}
		_, err := u.updateAvatarIfProvided(context.Background(), 7, user, bytes.NewReader(nil), 1, "image/png")
		if err != domain.ErrInvalidAvatar {
			t.Fatalf("expected ErrInvalidAvatar, got %v", err)
		}
	})
}
