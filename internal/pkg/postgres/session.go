package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	db *Postgres
}

func NewSessionRepo(db *Postgres) *SessionRepo {
	return &SessionRepo{db: db}
}

func (s *SessionRepo) SaveSession(ctx context.Context, userId uuid.UUID, refreshToken string, expiresAt time.Time) error {
	delete_sql := "delete from user_session where user_id=$1"
	insert_sql :=
		`insert into user_session (user_id, refresh_token, expires_at) 
values ($1, $2, $3)`

	_, err := s.db.Pool.Exec(ctx, delete_sql, userId)

	if err != nil {
		return fmt.Errorf("fail to save session: %w", err)
	}

	_, err = s.db.Pool.Exec(ctx, insert_sql, userId, refreshToken, expiresAt)

	if err != nil {
		return fmt.Errorf("fail to save session: %w", err)
	}

	return nil
}

// возвращается просто строка, а не token пара, т. к. access в бд не храним, в сессии он не нужен.
func (s *SessionRepo) GetSession(ctx context.Context, userId uuid.UUID) (string, error) {
	sql := "select refresh_token, expires_at from user_session where user_id=$1"

	var refreshToken string
	var expiresAt time.Time

	err := s.db.Pool.QueryRow(ctx, sql, userId).Scan(&refreshToken, &expiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNoSession
		}
		return "", fmt.Errorf("session query failed: %w", err)
	}

	if time.Now().After(expiresAt) {
		_ = s.DeleteSession(ctx, userId)
		return "", domain.ErrNoSession
	}

	return refreshToken, nil
}

func (s *SessionRepo) DeleteSession(ctx context.Context, userId uuid.UUID) error {
	delete_sql := "delete from user_session where user_id=$1"

	_, err := s.db.Pool.Exec(ctx, delete_sql, userId)

	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	// if result.RowsAffected() == 0 {
	// 	return domain.ErrNoSession
	// } // наверное, не нужно

	return nil
}
