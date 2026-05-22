package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
)

type stubAvatarStore struct {
	presign func(ctx context.Context, key string, ttl time.Duration) (string, error)
}

const testAvatarKey = "users/42/avatar/current.png"

var errPresignFailed = errors.New("presign failed")

func (s stubAvatarStore) PutObject(context.Context, string, io.Reader, int64, string) error {
	return nil
}

func (s stubAvatarStore) DeleteObject(context.Context, string) error {
	return nil
}

func (s stubAvatarStore) PresignGetObject(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if s.presign != nil {
		return s.presign(ctx, key, ttl)
	}

	return "", nil
}

func (s stubAvatarStore) GetObject(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}

func TestUpdateProfile_IgnoresAvatarPresignFailureAfterBirthdateUpdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := time.Date(1991, 2, 3, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: ptrToTestAvatarKey(),
		}, nil)

	repo.EXPECT().
		UpdateBirthdate(gomock.Any(), int64(42), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int64, got *time.Time) (*domain.User, error) {
			if got == nil || !got.Equal(newBirthdate) {
				t.Fatalf("birthdate = %v, want %v", got, newBirthdate)
			}

			return &domain.User{
				ID:            42,
				Email:         "user@example.com",
				Role:          "user",
				Birthdate:     &newBirthdate,
				AvatarFileKey: ptrToTestAvatarKey(),
			}, nil
		})

	u := NewUserUsecase(
		repo,
		stubAvatarStore{
			presign: func(context.Context, string, time.Duration) (string, error) {
				return "", errPresignFailed
			},
		},
		nil,
	)

	resp, err := u.UpdateProfile(context.Background(), 42, "1991-02-03", nil, 0, "")
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	if resp.Birthdate == nil || *resp.Birthdate != "1991-02-03" {
		t.Fatalf("birthdate = %v, want %q", resp.Birthdate, "1991-02-03")
	}

	if resp.AvatarURL != "" {
		t.Fatalf("avatar_url = %q, want empty", resp.AvatarURL)
	}
}

func TestUpdateProfile_IgnoresNullAvatarPayload(t *testing.T) {
	t.Parallel()
	assertAvatarPayloadIgnored(t, "2000-10-07", []byte("null"), "")
}

func TestUpdateProfile_IgnoresNullAvatarPayloadWithImageContentType(t *testing.T) {
	t.Parallel()
	assertAvatarPayloadIgnored(t, "2004-10-07", []byte("null"), "image/png")
}

func ptrToTestAvatarKey() *string {
	key := testAvatarKey

	return &key
}

func TestUpdateProfile_IgnoresUnsupportedAvatarPayloadWithImageContentType(t *testing.T) {
	t.Parallel()
	assertAvatarPayloadIgnored(t, "2005-03-05", []byte("garbage-avatar-payload"), "image/png")
}

func assertAvatarPayloadIgnored(t *testing.T, birthdate string, avatarPayload []byte, contentType string) {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := mustParseBirthdate(t, birthdate)

	expectProfileBirthdateUpdate(t, repo, oldBirthdate, newBirthdate)

	u := NewUserUsecase(repo, stubAvatarStore{}, nil)

	resp, err := u.UpdateProfile(
		context.Background(),
		42,
		birthdate,
		bytes.NewReader(avatarPayload),
		int64(len(avatarPayload)),
		contentType,
	)
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	if resp.Birthdate == nil || *resp.Birthdate != birthdate {
		t.Fatalf("birthdate = %v, want %q", resp.Birthdate, birthdate)
	}
}

func expectProfileBirthdateUpdate(t *testing.T, repo *mocks.MockUserRepo, oldBirthdate, newBirthdate time.Time) {
	t.Helper()

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: ptrToTestAvatarKey(),
		}, nil)

	repo.EXPECT().
		UpdateBirthdate(gomock.Any(), int64(42), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int64, got *time.Time) (*domain.User, error) {
			if got == nil || !got.Equal(newBirthdate) {
				t.Fatalf("birthdate = %v, want %v", got, newBirthdate)
			}

			return &domain.User{
				ID:            42,
				Email:         "user@example.com",
				Role:          "user",
				Birthdate:     &newBirthdate,
				AvatarFileKey: ptrToTestAvatarKey(),
			}, nil
		})
}

func mustParseBirthdate(t *testing.T, value string) time.Time {
	t.Helper()

	birthdate, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("time.Parse(%q): %v", value, err)
	}

	return birthdate
}
