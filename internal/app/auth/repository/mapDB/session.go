package mapDB

import "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"

type SessionRepo struct {
	db *DB
}

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// При необходимости переименовать методы в SessionRepo интерфейс в repository/repository.go

func (s *SessionRepo) SaveSession(email string, tokens domain.TokenPair) error {
	return nil
}

func (s *SessionRepo) GetSession(email string) (domain.TokenPair, error) {
	return domain.TokenPair{}, nil
}

func (s *SessionRepo) DeleteSession(email string) error { return nil }
