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

func TestNewSessionRepo(t *testing.T) {
	t.Parallel()

	repo := NewSessionRepo(&Client{})
	if repo == nil || repo.db == nil {
		t.Fatal("expected repo with db")
	}
}

func TestSessionRepo_SaveSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expiresAt := time.Now().Add(time.Hour)
	pool := NewMockPool(ctrl)
	pool.EXPECT().
		Exec(gomock.Any(), sqlSaveSession, int64(1), "refresh", expiresAt).
		Return(pgconn.NewCommandTag("INSERT 1"), nil)

	repo := NewSessionRepo(&Client{Pool: pool})
	if err := repo.SaveSession(context.Background(), 1, "refresh", expiresAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionRepo_GetSession(t *testing.T) {
	t.Parallel()

	t.Run("no rows", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(2)...).Return(pgx.ErrNoRows)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetSession, int64(1)).Return(row)

		repo := NewSessionRepo(&Client{Pool: pool})
		_, err := repo.GetSession(context.Background(), 1)
		if !errors.Is(err, userdomain.ErrNoSession) {
			t.Fatalf("expected ErrNoSession, got %v", err)
		}
	})

	t.Run("expired session", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expiredAt := time.Now().Add(-time.Hour)
		row := NewMockRow(ctrl)
		row.EXPECT().
			Scan(anyArgs(2)...).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "refresh"
				*dest[1].(*time.Time) = expiredAt

				return nil
			})

		pool := NewMockPool(ctrl)
		gomock.InOrder(
			pool.EXPECT().QueryRow(gomock.Any(), sqlGetSession, int64(1)).Return(row),
			pool.EXPECT().Exec(gomock.Any(), sqlDeleteSession, int64(1)).Return(pgconn.NewCommandTag("DELETE 1"), nil),
		)

		repo := NewSessionRepo(&Client{Pool: pool})
		_, err := repo.GetSession(context.Background(), 1)
		if !errors.Is(err, userdomain.ErrNoSession) {
			t.Fatalf("expected ErrNoSession, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expiresAt := time.Now().Add(time.Hour)
		row := NewMockRow(ctrl)
		row.EXPECT().
			Scan(anyArgs(2)...).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "refresh"
				*dest[1].(*time.Time) = expiresAt

				return nil
			})

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetSession, int64(1)).Return(row)

		repo := NewSessionRepo(&Client{Pool: pool})
		got, err := repo.GetSession(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "refresh" {
			t.Fatalf("expected refresh token %q, got %q", "refresh", got)
		}
	})
}

func TestSessionRepo_DeleteSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pool := NewMockPool(ctrl)
	pool.EXPECT().
		Exec(gomock.Any(), sqlDeleteSession, int64(1)).
		Return(pgconn.NewCommandTag("DELETE 1"), nil)

	repo := NewSessionRepo(&Client{Pool: pool})
	if err := repo.DeleteSession(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
