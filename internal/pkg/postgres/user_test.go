package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	userdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"
)

func TestNewUserRepo(t *testing.T) {
	t.Parallel()

	repo := NewUserRepo(&Client{})
	if repo == nil || repo.db == nil {
		t.Fatal("expected repo with db")
	}
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	birthdate := now.Add(-24 * time.Hour)
	avatar := "avatars/user.png"
	user := userdomain.User{
		ID:               1,
		Email:            "user@example.com",
		Password:         "hash",
		Birthdate:        &birthdate,
		AvatarFileKey:    &avatar,
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		expectUserRowScan(row, user)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetUserByEmail, "user@example.com").Return(row)

		repo := NewUserRepo(&Client{Pool: pool})
		got, err := repo.GetUserByEmail(context.Background(), "user@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Email != user.Email || got.ID != user.ID {
			t.Fatalf("unexpected user: %#v", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(9)...).Return(pgx.ErrNoRows)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetUserByEmail, "missing@example.com").Return(row)

		repo := NewUserRepo(&Client{Pool: pool})
		_, err := repo.GetUserByEmail(context.Background(), "missing@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUserRepo_GetUserByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	row := NewMockRow(ctrl)
	row.EXPECT().Scan(anyArgs(9)...).Return(pgx.ErrNoRows)

	pool := NewMockPool(ctrl)
	pool.EXPECT().QueryRow(gomock.Any(), sqlGetUserByID, int64(42)).Return(row)

	repo := NewUserRepo(&Client{Pool: pool})
	_, err := repo.GetUserByID(context.Background(), 42)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_CreateUser(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	user := userdomain.User{
		ID:               1,
		Email:            "user@example.com",
		Password:         "hash",
		RegistrationDate: now,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	t.Run("duplicate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(9)...).Return(&pgconn.PgError{Code: "23505"})

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlCreateUser, "user@example.com", "hash").Return(row)

		repo := NewUserRepo(&Client{Pool: pool})
		_, err := repo.CreateUser(context.Background(), "user@example.com", "hash")
		if !errors.Is(err, ErrUserAlreadyExists) {
			t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		expectUserRowScan(row, user)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlCreateUser, "user@example.com", "hash").Return(row)

		repo := NewUserRepo(&Client{Pool: pool})
		got, err := repo.CreateUser(context.Background(), "user@example.com", "hash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Email != user.Email {
			t.Fatalf("expected email %q, got %q", user.Email, got.Email)
		}
	})
}

func TestUserRepo_UpdateUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	row := NewMockRow(ctrl)
	row.EXPECT().Scan(anyArgs(9)...).Return(pgx.ErrNoRows)

	pool := NewMockPool(ctrl)
	pool.EXPECT().QueryRow(gomock.Any(), sqlUpdateUser, "hash", gomock.Any(), "user@example.com").Return(row)

	repo := NewUserRepo(&Client{Pool: pool})
	_, err := repo.UpdateUser(context.Background(), "user@example.com", "hash")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_DeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool(ctrl)
		pool.EXPECT().
			Exec(gomock.Any(), sqlDeleteUser, "user@example.com").
			Return(pgconn.NewCommandTag("DELETE 0"), nil)

		repo := NewUserRepo(&Client{Pool: pool})
		err := repo.DeleteUser(context.Background(), "user@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool(ctrl)
		pool.EXPECT().
			Exec(gomock.Any(), sqlDeleteUser, "user@example.com").
			Return(pgconn.NewCommandTag("DELETE 1"), nil)

		repo := NewUserRepo(&Client{Pool: pool})
		if err := repo.DeleteUser(context.Background(), "user@example.com"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUserRepo_UpdateBirthdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	row := NewMockRow(ctrl)
	row.EXPECT().Scan(anyArgs(9)...).Return(pgx.ErrNoRows)

	pool := NewMockPool(ctrl)
	pool.EXPECT().QueryRow(gomock.Any(), sqlUpdateUserBirthdate, gomock.Any(), int64(7)).Return(row)

	repo := NewUserRepo(&Client{Pool: pool})
	_, err := repo.UpdateBirthdate(context.Background(), 7, nil)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_UpdateAvatarFileKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "avatars/new.png"
	row := NewMockRow(ctrl)
	row.EXPECT().Scan(anyArgs(9)...).Return(pgx.ErrNoRows)

	pool := NewMockPool(ctrl)
	pool.EXPECT().QueryRow(gomock.Any(), sqlUpdateUserAvatarFileKey, &key, int64(7)).Return(row)

	repo := NewUserRepo(&Client{Pool: pool})
	_, err := repo.UpdateAvatarFileKey(context.Background(), 7, &key)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool(ctrl)
		pool.EXPECT().
			Exec(gomock.Any(), sqlUpdateUserPasswordByID, "hash", int64(5)).
			Return(pgconn.NewCommandTag("UPDATE 0"), nil)

		repo := NewUserRepo(&Client{Pool: pool})
		err := repo.UpdatePassword(context.Background(), 5, "hash")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool(ctrl)
		pool.EXPECT().
			Exec(gomock.Any(), sqlUpdateUserPasswordByID, "hash", int64(5)).
			Return(pgconn.NewCommandTag("UPDATE 1"), nil)

		repo := NewUserRepo(&Client{Pool: pool})
		if err := repo.UpdatePassword(context.Background(), 5, "hash"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
