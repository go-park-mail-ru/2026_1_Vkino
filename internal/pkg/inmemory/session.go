package inmemory

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
)

type SessionRepo struct {
	db *DB
}

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (s *SessionRepo) SaveSession(email string, tokens domain.TokenPair) error {
	data, err := serializer.Serialize(tokens)
	if err != nil {
		return err
	}

	_ = s.db.Delete("sessions", email)

	err = s.db.Save("sessions", email, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionRepo) GetSession(email string) (*domain.TokenPair, error) {
	data, err := s.db.Get("sessions", email)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrTableNotFound) {
			return &domain.TokenPair{}, domain.ErrNoSession
		}

		return &domain.TokenPair{}, err
	}

	var tokens domain.TokenPair
	if err := serializer.Deserialize(data, &tokens); err != nil {
		return &domain.TokenPair{}, err
	}

	return &tokens, nil
}

func (s *SessionRepo) DeleteSession(email string) error {
	err := s.db.Delete("sessions", email)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrTableNotFound) {
			return domain.ErrNoSession
		}

		return err
	}

	return nil
}
