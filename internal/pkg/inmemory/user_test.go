package inmemory

import (
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
)

type userRepoTestModel string

func (m userRepoTestModel) Name() string {
	return string(m)
}

func newTestUserRepo(withUsersTable bool) *UserRepo {
	models := []Named{}
	if withUsersTable {
		models = append(models, userRepoTestModel("users"))
	}

	db := NewDB(models)

	return NewUserRepo(db)
}

func usersByEmail(users []*domain.User) map[string]*domain.User {
	result := make(map[string]*domain.User, len(users))
	for _, user := range users {
		result[user.Email] = user
	}

	return result
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withTable  bool
		email      string
		prepare    func(t *testing.T, repo *UserRepo)
		wantEmail  string
		wantErrIs  error
		wantAnyErr bool
	}{
		{
			name:      "success",
			withTable: true,
			email:     "user@example.com",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				_, err := repo.CreateUser("user@example.com", "hash-password")
				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}
			},
			wantEmail: "user@example.com",
		},
		{
			name:      "user not found",
			withTable: true,
			email:     "missing@example.com",
			wantErrIs: ErrUserNotFound,
		},
		{
			name:      "table not found maps to user not found",
			withTable: false,
			email:     "user@example.com",
			wantErrIs: ErrUserNotFound,
		},
		{
			name:      "deserialize error",
			withTable: true,
			email:     "broken@example.com",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				err := repo.db.Save("users", "broken@example.com", []byte("not valid json"))
				if err != nil {
					t.Fatalf("save broken user: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetUserByEmail(tt.email)

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

			if got == nil {
				t.Fatal("expected non-nil user")
			}

			if got.Email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, got.Email)
			}
		})
	}
}

func TestUserRepo_GetUserByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withTable  bool
		prepare    func(t *testing.T, repo *UserRepo) int64
		wantErrIs  error
		wantAnyErr bool
		wantEmail  string
	}{
		{
			name:      "success",
			withTable: true,
			prepare: func(t *testing.T, repo *UserRepo) int64 {
				t.Helper()

				user, err := repo.CreateUser("user@example.com", "hash-password")
				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}

				return user.ID
			},
			wantEmail: "user@example.com",
		},
		{
			name:      "user not found",
			withTable: true,
			prepare: func(t *testing.T, repo *UserRepo) int64 {
				t.Helper()

				return 99999
			},
			wantErrIs: ErrUserNotFound,
		},
		{
			name:      "table not found maps to user not found",
			withTable: false,
			prepare: func(t *testing.T, repo *UserRepo) int64 {
				t.Helper()

				return 99999
			},
			wantErrIs: ErrUserNotFound,
		},
		{
			name:      "deserialize error",
			withTable: true,
			prepare: func(t *testing.T, repo *UserRepo) int64 {
				t.Helper()

				err := repo.db.Save("users", "broken@example.com", []byte("not valid json"))
				if err != nil {
					t.Fatalf("save broken user: %v", err)
				}

				return 99999
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)
			id := tt.prepare(t, repo)

			got, err := repo.GetUserByID(id)

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

			if got == nil {
				t.Fatal("expected non-nil user")
			}

			if got.ID != id {
				t.Fatalf("expected id %v, got %v", id, got.ID)
			}

			if got.Email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, got.Email)
			}
		})
	}
}

func TestUserRepo_CreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		withTable    bool
		login        string
		password     string
		prepare      func(t *testing.T, repo *UserRepo)
		wantErrIs    error
		wantAnyErr   bool
		wantEmail    string
		wantPassword string
	}{
		{
			name:         "success",
			withTable:    true,
			login:        "user@example.com",
			password:     "hash-password",
			wantEmail:    "user@example.com",
			wantPassword: "hash-password",
		},
		{
			name:      "already exists",
			withTable: true,
			login:     "user@example.com",
			password:  "new-password",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				_, err := repo.CreateUser("user@example.com", "old-password")
				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}
			},
			wantErrIs: ErrUserAlreadyExists,
		},
		{
			name:      "table not found",
			withTable: false,
			login:     "user@example.com",
			password:  "hash-password",
			wantErrIs: ErrTableNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			before := time.Now()
			got, err := repo.CreateUser(tt.login, tt.password)
			after := time.Now()

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

			if got == nil {
				t.Fatal("expected non-nil user")
			}

			if got.ID == 0 {
				t.Fatal("expected non-zero id")
			}

			if got.Email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, got.Email)
			}

			if got.Password != tt.wantPassword {
				t.Fatalf("expected password %q, got %q", tt.wantPassword, got.Password)
			}

			if !got.IsActive {
				t.Fatal("expected IsActive=true")
			}

			if got.RegistrationDate.Before(before) || got.RegistrationDate.After(after) {
				t.Fatalf("unexpected registration date: %v", got.RegistrationDate)
			}

			if got.CreatedAt.Before(before) || got.CreatedAt.After(after) {
				t.Fatalf("unexpected created_at: %v", got.CreatedAt)
			}

			if got.UpdatedAt.Before(before) || got.UpdatedAt.After(after) {
				t.Fatalf("unexpected updated_at: %v", got.UpdatedAt)
			}

			stored, err := repo.GetUserByEmail(tt.login)
			if err != nil {
				t.Fatalf("get created user: %v", err)
			}

			if stored.Email != got.Email {
				t.Fatalf("expected stored email %q, got %q", got.Email, stored.Email)
			}
		})
	}
}

func TestUserRepo_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		withTable       bool
		login           string
		newPassword     string
		prepare         func(t *testing.T, repo *UserRepo)
		wantErrIs       error
		wantAnyErr      bool
		wantPassword    string
		wantUpdatedDiff bool
	}{
		{
			name:        "success",
			withTable:   true,
			login:       "user@example.com",
			newPassword: "new-password",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				_, err := repo.CreateUser("user@example.com", "old-password")
				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}
			},
			wantPassword:    "new-password",
			wantUpdatedDiff: true,
		},
		{
			name:        "user not found",
			withTable:   true,
			login:       "missing@example.com",
			newPassword: "new-password",
			wantErrIs:   ErrUserNotFound,
		},
		{
			name:        "table not found maps to user not found",
			withTable:   false,
			login:       "user@example.com",
			newPassword: "new-password",
			wantErrIs:   ErrUserNotFound,
		},
		{
			name:        "deserialize error on get user",
			withTable:   true,
			login:       "broken@example.com",
			newPassword: "new-password",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				err := repo.db.Save("users", "broken@example.com", []byte("not valid json"))
				if err != nil {
					t.Fatalf("save broken user: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			var oldUser *domain.User
			if tt.withTable && tt.login == "user@example.com" {
				oldUser, _ = repo.GetUserByEmail(tt.login)
			}

			got, err := repo.UpdateUser(tt.login, tt.newPassword)

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

			if got == nil {
				t.Fatal("expected non-nil user")
			}

			if got.Password != tt.wantPassword {
				t.Fatalf("expected password %q, got %q", tt.wantPassword, got.Password)
			}

			stored, err := repo.GetUserByEmail(tt.login)
			if err != nil {
				t.Fatalf("get updated user: %v", err)
			}

			if stored.Password != tt.wantPassword {
				t.Fatalf("expected stored password %q, got %q", tt.wantPassword, stored.Password)
			}

			if tt.wantUpdatedDiff && oldUser != nil && !stored.UpdatedAt.After(oldUser.UpdatedAt) &&
				!stored.UpdatedAt.Equal(oldUser.UpdatedAt) {
				t.Fatalf("expected updated_at to be >= old updated_at, old=%v new=%v", oldUser.UpdatedAt, stored.UpdatedAt)
			}
		})
	}
}

func TestUserRepo_GetAllUsers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withTable  bool
		prepare    func(t *testing.T, repo *UserRepo)
		wantLen    int
		wantEmails []string
		wantAnyErr bool
	}{
		{
			name:      "returns only active users",
			withTable: true,
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				activeUser, err := repo.CreateUser("active@example.com", "pass-1")
				if err != nil {
					t.Fatalf("create active user: %v", err)
				}

				inactiveUser, err := repo.CreateUser("inactive@example.com", "pass-2")
				if err != nil {
					t.Fatalf("create inactive user: %v", err)
				}

				inactiveUser.IsActive = false

				data, err := serializer.Serialize(*inactiveUser)
				if err != nil {
					t.Fatalf("serialize inactive user: %v", err)
				}

				if err := repo.db.Update("users", inactiveUser.Email, data); err != nil {
					t.Fatalf("update inactive user: %v", err)
				}

				if activeUser.Email != "active@example.com" {
					t.Fatalf("unexpected active user email %q", activeUser.Email)
				}
			},
			wantLen:    1,
			wantEmails: []string{"active@example.com"},
		},
		{
			name:      "table not found returns empty slice",
			withTable: false,
			wantLen:   0,
		},
		{
			name:      "deserialize error",
			withTable: true,
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				err := repo.db.Save("users", "broken@example.com", []byte("not valid json"))
				if err != nil {
					t.Fatalf("save broken user: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetAllUsers()

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf("expected %d users, got %d", tt.wantLen, len(got))
			}

			if len(tt.wantEmails) > 0 {
				gotMap := usersByEmail(got)
				for _, email := range tt.wantEmails {
					if _, ok := gotMap[email]; !ok {
						t.Fatalf("expected user %q in result", email)
					}
				}
			}
		})
	}
}

func TestUserRepo_DeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		withTable bool
		login     string
		prepare   func(t *testing.T, repo *UserRepo)
		wantErrIs error
	}{
		{
			name:      "success",
			withTable: true,
			login:     "user@example.com",
			prepare: func(t *testing.T, repo *UserRepo) {
				t.Helper()

				_, err := repo.CreateUser("user@example.com", "hash-password")
				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}
			},
		},
		{
			name:      "user not found",
			withTable: true,
			login:     "missing@example.com",
			wantErrIs: ErrUserNotFound,
		},
		{
			name:      "table not found maps to user not found",
			withTable: false,
			login:     "user@example.com",
			wantErrIs: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(tt.withTable)

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			err := repo.DeleteUser(tt.login)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, err = repo.GetUserByEmail(tt.login)
			if !errors.Is(err, ErrUserNotFound) {
				t.Fatalf("expected error %v after delete, got %v", ErrUserNotFound, err)
			}
		})
	}
}
