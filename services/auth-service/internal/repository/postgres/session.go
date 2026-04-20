package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	db *corepostgres.Client
}

func NewSessionRepo(db *corepostgres.Client) *SessionRepo {
	return &SessionRepo{db: db}
}

func (s *SessionRepo) SaveSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, sqlSaveSession, userID, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("fail to save session: %w", err)
	}

	return nil
}

func (s *SessionRepo) GetSession(ctx context.Context, userID int64) (string, error) {
	var refreshToken string
	
	var expiresAt time.Time

	err := s.db.QueryRow(ctx, sqlGetSession, userID).Scan(&refreshToken, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNoSession
		}

		return "", fmt.Errorf("session query failed: %w", err)
	}

	if time.Now().After(expiresAt) {
		_ = s.DeleteSession(ctx, userID)

		return "", domain.ErrNoSession
	}

	return refreshToken, nil
}

func (s *SessionRepo) DeleteSession(ctx context.Context, userID int64) error {
	_, err := s.db.Exec(ctx, sqlDeleteSession, userID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}
