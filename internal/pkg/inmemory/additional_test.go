package inmemory

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
)

func TestDB_TableExists(t *testing.T) {
	t.Parallel()

	db := NewDB([]Named{testModel("movies")})
	if !db.TableExists("movies") {
		t.Fatal("expected table movies to exist")
	}

	if db.TableExists("actors") {
		t.Fatal("expected table actors to be absent")
	}
}

func TestMovieRepo_GetMovieByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int64
		prepare    func(t *testing.T, repo *MovieRepo)
		wantTitle  string
		wantErrIs  error
		wantAnyErr bool
	}{
		{
			name:      "success",
			id:        101,
			wantTitle: "Дюна: Часть Вторая",
		},
		{
			name:      "not found",
			id:        999,
			wantErrIs: ErrMovieNotFound,
		},
		{
			name: "deserialize error",
			id:   999,
			prepare: func(t *testing.T, repo *MovieRepo) {
				t.Helper()

				if err := repo.db.Save("movies", "999", []byte("broken")); err != nil {
					t.Fatalf("save broken movie: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := NewDB([]Named{testModel("selections"), testModel("movies"), testModel("actors")})
			repo := NewMovieRepo(db)
			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetMovieByID(tt.id)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return
			}

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected non-nil error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Title != tt.wantTitle {
				t.Fatalf("expected title %q, got %q", tt.wantTitle, got.Title)
			}
		})
	}
}

func TestMovieRepo_GetActorByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int64
		prepare    func(t *testing.T, repo *MovieRepo)
		wantName   string
		wantErrIs  error
		wantAnyErr bool
	}{
		{
			name:     "success",
			id:       1,
			wantName: "Тимати Шаламе",
		},
		{
			name:      "not found",
			id:        999,
			wantErrIs: ErrActorNotFound,
		},
		{
			name: "deserialize error",
			id:   999,
			prepare: func(t *testing.T, repo *MovieRepo) {
				t.Helper()

				if err := repo.db.Save("actors", "999", []byte("broken")); err != nil {
					t.Fatalf("save broken actor: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := NewDB([]Named{testModel("selections"), testModel("movies"), testModel("actors")})
			repo := NewMovieRepo(db)
			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetActorByID(tt.id)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return
			}

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected non-nil error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.FullName != tt.wantName {
				t.Fatalf("expected full name %q, got %q", tt.wantName, got.FullName)
			}
		})
	}
}

func TestMovieRepo_GetMovieByID_CustomSerialized(t *testing.T) {
	t.Parallel()

	db := NewDB([]Named{testModel("selections"), testModel("movies"), testModel("actors")})
	repo := NewMovieRepo(db)

	custom := domain.MovieResponse{ID: 777, Title: "Custom"}
	data, err := serializer.Serialize(custom)
	if err != nil {
		t.Fatalf("serialize custom movie: %v", err)
	}

	if err = repo.db.Save("movies", "777", data); err != nil {
		t.Fatalf("save custom movie: %v", err)
	}

	got, err := repo.GetMovieByID(777)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title != "Custom" {
		t.Fatalf("expected title %q, got %q", "Custom", got.Title)
	}
}
