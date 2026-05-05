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
	return nil, nil
}

func TestUpdateProfile_IgnoresAvatarPresignFailureAfterBirthdateUpdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	avatarKey := "users/42/avatar/current.png"
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := time.Date(1991, 2, 3, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: &avatarKey,
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
				AvatarFileKey: &avatarKey,
			}, nil
		})

	u := NewUserUsecase(
		repo,
		stubAvatarStore{
			presign: func(context.Context, string, time.Duration) (string, error) {
				return "", errors.New("presign failed")
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	avatarKey := "users/42/avatar/current.png"
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := time.Date(2000, 10, 7, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: &avatarKey,
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
				AvatarFileKey: &avatarKey,
			}, nil
		})

	u := NewUserUsecase(repo, stubAvatarStore{}, nil)

	resp, err := u.UpdateProfile(
		context.Background(),
		42,
		"2000-10-07",
		bytes.NewReader([]byte("null")),
		int64(len("null")),
		"",
	)
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	if resp.Birthdate == nil || *resp.Birthdate != "2000-10-07" {
		t.Fatalf("birthdate = %v, want %q", resp.Birthdate, "2000-10-07")
	}
}

func TestUpdateProfile_IgnoresNullAvatarPayloadWithImageContentType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	avatarKey := "users/42/avatar/current.png"
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := time.Date(2004, 10, 7, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: &avatarKey,
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
				AvatarFileKey: &avatarKey,
			}, nil
		})

	u := NewUserUsecase(repo, stubAvatarStore{}, nil)

	resp, err := u.UpdateProfile(
		context.Background(),
		42,
		"2004-10-07",
		bytes.NewReader([]byte("null")),
		int64(len("null")),
		"image/png",
	)
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	if resp.Birthdate == nil || *resp.Birthdate != "2004-10-07" {
		t.Fatalf("birthdate = %v, want %q", resp.Birthdate, "2004-10-07")
	}
}

func TestUpdateProfile_IgnoresUnsupportedAvatarPayloadWithImageContentType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	avatarKey := "users/42/avatar/current.png"
	oldBirthdate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	newBirthdate := time.Date(2005, 3, 5, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&domain.User{
			ID:            42,
			Email:         "user@example.com",
			Role:          "user",
			Birthdate:     &oldBirthdate,
			AvatarFileKey: &avatarKey,
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
				AvatarFileKey: &avatarKey,
			}, nil
		})

	u := NewUserUsecase(repo, stubAvatarStore{}, nil)

	resp, err := u.UpdateProfile(
		context.Background(),
		42,
		"2005-03-05",
		bytes.NewReader([]byte("garbage-avatar-payload")),
		int64(len("garbage-avatar-payload")),
		"image/png",
	)
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}

	if resp.Birthdate == nil || *resp.Birthdate != "2005-03-05" {
		t.Fatalf("birthdate = %v, want %q", resp.Birthdate, "2005-03-05")
	}
}
