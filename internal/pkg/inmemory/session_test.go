package inmemory

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
)

type sessionTestModel string

func (m sessionTestModel) Name() string {
	return string(m)
}

func newTestSessionRepo(withSessionsTable bool) *SessionRepo {
	models := []Named{}
	if withSessionsTable {
		models = append(models, sessionTestModel("sessions"))
	}

	db := NewDB(models)
	return NewSessionRepo(db)
}

func TestSessionRepo_SaveSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		withTable      bool
		email          string
		initialTokens  *domain.TokenPair
		saveTokens     domain.TokenPair
		wantErrIs      error
		expectedStored *domain.TokenPair
	}{
		{
			name:      "save new session",
			withTable: true,
			email:     "user@example.com",
			saveTokens: domain.TokenPair{
				AccessToken:  "access-1",
				RefreshToken: "refresh-1",
			},
			expectedStored: &domain.TokenPair{
				AccessToken:  "access-1",
				RefreshToken: "refresh-1",
			},
		},
		{
			name:      "overwrite existing session",
			withTable: true,
			email:     "user@example.com",
			initialTokens: &domain.TokenPair{
				AccessToken:  "old-access",
				RefreshToken: "old-refresh",
			},
			saveTokens: domain.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			},
			expectedStored: &domain.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			},
		},
		{
			name:      "table not found",
			withTable: false,
			email:     "user@example.com",
			saveTokens: domain.TokenPair{
				AccessToken:  "access-1",
				RefreshToken: "refresh-1",
			},
			wantErrIs: ErrTableNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestSessionRepo(tt.withTable)

			if tt.initialTokens != nil {
				if err := repo.SaveSession(tt.email, *tt.initialTokens); err != nil {
					t.Fatalf("prepare initial session: %v", err)
				}
			}

			err := repo.SaveSession(tt.email, tt.saveTokens)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := repo.GetSession(tt.email)
			if err != nil {
				t.Fatalf("get saved session: %v", err)
			}

			if tt.expectedStored != nil && *got != *tt.expectedStored {
				t.Fatalf("expected stored tokens %+v, got %+v", *tt.expectedStored, *got)
			}
		})
	}
}

func TestSessionRepo_GetSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withTable  bool
		email      string
		prepare    func(t *testing.T, repo *SessionRepo)
		want       *domain.TokenPair
		wantErrIs  error
		wantAnyErr bool
	}{
		{
			name:      "success",
			withTable: true,
			email:     "user@example.com",
			prepare: func(t *testing.T, repo *SessionRepo) {
				t.Helper()

				err := repo.SaveSession("user@example.com", domain.TokenPair{
					AccessToken:  "access-1",
					RefreshToken: "refresh-1",
				})
				if err != nil {
					t.Fatalf("save session: %v", err)
				}
			},
			want: &domain.TokenPair{
				AccessToken:  "access-1",
				RefreshToken: "refresh-1",
			},
		},
		{
			name:      "session not found",
			withTable: true,
			email:     "missing@example.com",
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:      "table not found maps to no session",
			withTable: false,
			email:     "user@example.com",
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:      "deserialize error",
			withTable: true,
			email:     "broken@example.com",
			prepare: func(t *testing.T, repo *SessionRepo) {
				t.Helper()

				err := repo.db.Save("sessions", "broken@example.com", []byte("not valid json"))
				if err != nil {
					t.Fatalf("save broken session: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestSessionRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetSession(tt.email)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if *got != *tt.want {
				t.Fatalf("expected tokens %+v, got %+v", *tt.want, *got)
			}
		})
	}
}

func TestSessionRepo_DeleteSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withTable  bool
		email      string
		prepare    func(t *testing.T, repo *SessionRepo)
		wantErrIs  error
		checkAfter bool
	}{
		{
			name:      "success",
			withTable: true,
			email:     "user@example.com",
			prepare: func(t *testing.T, repo *SessionRepo) {
				t.Helper()

				err := repo.SaveSession("user@example.com", domain.TokenPair{
					AccessToken:  "access-1",
					RefreshToken: "refresh-1",
				})
				if err != nil {
					t.Fatalf("save session: %v", err)
				}
			},
			checkAfter: true,
		},
		{
			name:      "session not found",
			withTable: true,
			email:     "missing@example.com",
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:      "table not found maps to no session",
			withTable: false,
			email:     "user@example.com",
			wantErrIs: domain.ErrNoSession,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestSessionRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			err := repo.DeleteSession(tt.email)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkAfter {
				_, err = repo.GetSession(tt.email)
				if !errors.Is(err, domain.ErrNoSession) {
					t.Fatalf("expected error %v after delete, got %v", domain.ErrNoSession, err)
				}
			}
		})
	}
}
