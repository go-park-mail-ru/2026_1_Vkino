package inmemory

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/serializer"
	"github.com/google/uuid"
)

type testModel string

func (m testModel) Name() string {
	return string(m)
}

func newTestMovieRepo() *MovieRepo {
	db := NewDB([]Named{
		testModel("selections"),
	})

	return NewMovieRepo(db)
}

func selectionsToMap(selections []domain.SelectionResponse) map[string]int {
	result := make(map[string]int, len(selections))
	for _, s := range selections {
		result[s.Title] = len(s.Movies)
	}

	return result
}

func TestMovieRepo_GetSelectionByTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		title      string
		prepare    func(t *testing.T, repo *MovieRepo)
		wantTitle  string
		wantMovies int
		wantErrIs  error
		wantAnyErr bool
	}{
		{
			name:       "success popular",
			title:      "popular",
			wantTitle:  "Популярные",
			wantMovies: 10,
		},
		{
			name:       "success new",
			title:      "new",
			wantTitle:  "Новинки",
			wantMovies: 10,
		},
		{
			name:       "success top",
			title:      "top",
			wantTitle:  "Топ-10",
			wantMovies: 10,
		},
		{
			name:      "selection not found",
			title:     "unknown",
			wantErrIs: ErrSelectionNotFound,
		},
		{
			name:  "deserialize error",
			title: "broken",
			prepare: func(t *testing.T, repo *MovieRepo) {
				t.Helper()

				err := repo.db.Save("selections", "broken", []byte("not valid json"))
				if err != nil {
					t.Fatalf("failed to save broken data: %v", err)
				}
			},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newTestMovieRepo()

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetSelectionByTitle(tt.title)

			switch {
			case tt.wantErrIs != nil:
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return

			case tt.wantAnyErr:
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}

				return

			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if got.Title != tt.wantTitle {
				t.Fatalf("expected title %q, got %q", tt.wantTitle, got.Title)
			}

			if len(got.Movies) != tt.wantMovies {
				t.Fatalf("expected %d movies, got %d", tt.wantMovies, len(got.Movies))
			}
		})
	}
}

func TestMovieRepo_GetAllSelections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prepare   func(t *testing.T, repo *MovieRepo)
		wantCount int
		wantMap   map[string]int
	}{
		{
			name:      "default selections",
			wantCount: 3,
			wantMap: map[string]int{
				"Популярные": 10,
				"Новинки":    10,
				"Топ-10":     10,
			},
		},
		{
			name: "broken selection is skipped",
			prepare: func(t *testing.T, repo *MovieRepo) {
				t.Helper()

				err := repo.db.Save("selections", "broken", []byte("invalid json"))
				if err != nil {
					t.Fatalf("failed to save broken selection: %v", err)
				}
			},
			wantCount: 3,
			wantMap: map[string]int{
				"Популярные": 10,
				"Новинки":    10,
				"Топ-10":     10,
			},
		},
		{
			name: "extra valid selection is returned",
			prepare: func(t *testing.T, repo *MovieRepo) {
				t.Helper()

				custom := domain.SelectionResponse{
					Title: "Классика",
					Movies: []domain.MoviePreview{
						{
							ID:             uuid.New(),
							Title:          "The Godfather",
							PictureFileKey: "img/godfather.jpg",
						},
					},
				}

				data, err := serializer.Serialize(custom)
				if err != nil {
					t.Fatalf("failed to serialize custom selection: %v", err)
				}

				err = repo.db.Save("selections", "classic", data)
				if err != nil {
					t.Fatalf("failed to save custom selection: %v", err)
				}
			},
			wantCount: 4,
			wantMap: map[string]int{
				"Популярные": 10,
				"Новинки":    10,
				"Топ-10":     10,
				"Классика":   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newTestMovieRepo()

			if tt.prepare != nil {
				tt.prepare(t, repo)
			}

			got, err := repo.GetAllSelections()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Fatalf("expected %d selections, got %d", tt.wantCount, len(got))
			}

			gotMap := selectionsToMap(got)
			if !reflect.DeepEqual(gotMap, tt.wantMap) {
				t.Fatalf("unexpected selections map:\nwant: %#v\ngot:  %#v", tt.wantMap, gotMap)
			}
		})
	}
}
