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
	db *Client
}

func NewSessionRepo(db *Client) *SessionRepo {
	return &SessionRepo{db: db}
}

func (s *SessionRepo) SaveSession(ctx context.Context, userId uuid.UUID, refreshToken string, expiresAt time.Time) error {
	_, err := s.db.Pool.Exec(ctx, sqlSaveSession, userId, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("fail to save session: %w", err)
	}

	return nil
}

func (s *SessionRepo) GetSession(ctx context.Context, userId uuid.UUID) (string, error) {
	var refreshToken string
	var expiresAt time.Time

	err := s.db.Pool.QueryRow(ctx, sqlGetSession, userId).Scan(&refreshToken, &expiresAt)

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
	_, err := s.db.Pool.Exec(ctx, sqlDeleteSession, userId)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}
