package postgres

import (
	"context"
	"errors"
	"testing"

	moviedomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"
)

func TestNewMovieRepo(t *testing.T) {
	t.Parallel()

	repo := NewMovieRepo(&Client{})
	if repo == nil || repo.db == nil {
		t.Fatal("expected repo with db")
	}
}

func TestMovieRepo_GetSelectionByTitle(t *testing.T) {
	t.Parallel()

	t.Run("query error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		pool := NewMockPool(ctrl)
		pool.EXPECT().Query(gomock.Any(), sqlGetSelectionByTitle, "popular").Return(nil, errors.New("query failed"))

		repo := NewMovieRepo(&Client{Pool: pool})
		_, err := repo.GetSelectionByTitle(context.Background(), "popular")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rows := NewMockRows(ctrl)
		expectSelectionMoviePreviewRows(rows, []moviedomain.MoviePreview{
			{ID: 1, Title: "Dune", ImgUrl: "img/1.jpg"},
			{ID: 2, Title: "Joker", ImgUrl: "img/2.jpg"},
		})

		pool := NewMockPool(ctrl)
		pool.EXPECT().Query(gomock.Any(), sqlGetSelectionByTitle, "popular").Return(rows, nil)

		repo := NewMovieRepo(&Client{Pool: pool})
		got, err := repo.GetSelectionByTitle(context.Background(), "popular")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Title != "popular" || len(got.Movies) != 2 {
			t.Fatalf("unexpected selection: %#v", got)
		}
	})
}

func TestMovieRepo_GetAllSelections(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	titleRows := NewMockRows(ctrl)
	expectStringRows(titleRows, []string{"popular", "new"})

	popularRows := NewMockRows(ctrl)
	expectSelectionMoviePreviewRows(popularRows, []moviedomain.MoviePreview{{ID: 1, Title: "Dune", ImgUrl: "img/1.jpg"}})

	newRows := NewMockRows(ctrl)
	expectSelectionMoviePreviewRows(newRows, []moviedomain.MoviePreview{{ID: 2, Title: "Joker", ImgUrl: "img/2.jpg"}})

	pool := NewMockPool(ctrl)
	gomock.InOrder(
		pool.EXPECT().Query(gomock.Any(), sqlGetAllSelectionTitles).Return(titleRows, nil),
		pool.EXPECT().Query(gomock.Any(), sqlGetSelectionByTitle, "popular").Return(popularRows, nil),
		pool.EXPECT().Query(gomock.Any(), sqlGetSelectionByTitle, "new").Return(newRows, nil),
	)

	repo := NewMovieRepo(&Client{Pool: pool})
	got, err := repo.GetAllSelections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 selections, got %d", len(got))
	}
}

func TestMovieRepo_GetMovieByID(t *testing.T) {
	t.Parallel()

	movie := moviedomain.MovieResponse{
		ID:                 1,
		Title:              "Dune",
		Description:        "desc",
		Director:           "director",
		ContentType:        "film",
		ReleaseYear:        2024,
		DurationSeconds:    120,
		AgeLimit:           16,
		OriginalLanguageID: 1,
		CountryID:          1,
		PictureFileKey:     "img/1.jpg",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	row := NewMockRow(ctrl)
	expectMovieRowScan(row, movie)

	genreRows := NewMockRows(ctrl)
	expectStringRows(genreRows, []string{"Drama", "Sci-Fi"})

	actorRows := NewMockRows(ctrl)
	expectActorPreviewRows(actorRows, []moviedomain.ActorPreview{{ID: 1, FullName: "Actor", PictureFileKey: "actor.jpg"}})

	pool := NewMockPool(ctrl)
	gomock.InOrder(
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetMovieByID, int64(1)).Return(row),
		pool.EXPECT().Query(gomock.Any(), sqlGetGenresByMovieID, int64(1)).Return(genreRows, nil),
		pool.EXPECT().Query(gomock.Any(), sqlGetActorsByMovieID, int64(1)).Return(actorRows, nil),
	)

	repo := NewMovieRepo(&Client{Pool: pool})
	got, err := repo.GetMovieByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Title != movie.Title || len(got.Genres) != 2 || len(got.Actors) != 1 {
		t.Fatalf("unexpected movie: %#v", got)
	}
}

func TestMovieRepo_GetActorByID(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(6)...).Return(pgx.ErrNoRows)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetActorByID, int64(1)).Return(row)

		repo := NewMovieRepo(&Client{Pool: pool})
		_, err := repo.GetActorByID(context.Background(), 1)
		if !errors.Is(err, ErrActorNotFound) {
			t.Fatalf("expected ErrActorNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		actor := moviedomain.ActorResponse{
			ID:             1,
			FullName:       "Actor",
			BirthDate:      "1990-01-01",
			Biography:      "bio",
			CountryID:      1,
			PictureFileKey: "actor.jpg",
		}

		row := NewMockRow(ctrl)
		expectActorRowScan(row, actor)

		movieRows := NewMockRows(ctrl)
		expectMoviePreviewRows(movieRows, []moviedomain.MoviePreview{{ID: 1, Title: "Dune", ImgUrl: "img/1.jpg"}})

		pool := NewMockPool(ctrl)
		gomock.InOrder(
			pool.EXPECT().QueryRow(gomock.Any(), sqlGetActorByID, int64(1)).Return(row),
			pool.EXPECT().Query(gomock.Any(), sqlGetMoviesByActorID, int64(1)).Return(movieRows, nil),
		)

		repo := NewMovieRepo(&Client{Pool: pool})
		got, err := repo.GetActorByID(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.FullName != actor.FullName || len(got.Movies) != 1 {
			t.Fatalf("unexpected actor: %#v", got)
		}
	})
}

func TestMovieRepo_GetEpisodesByMovieID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rows := NewMockRows(ctrl)
	expectEpisodeRows(rows, []moviedomain.EpisodeItemResponse{{ID: 1, MovieID: 1, Title: "Episode 1"}})

	pool := NewMockPool(ctrl)
	pool.EXPECT().Query(gomock.Any(), sqlGetEpisodesByMovieID, int64(1)).Return(rows, nil)

	repo := NewMovieRepo(&Client{Pool: pool})
	got, err := repo.GetEpisodesByMovieID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 || got[0].Title != "Episode 1" {
		t.Fatalf("unexpected episodes: %#v", got)
	}
}

func TestMovieRepo_GetEpisodePlayback(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(7)...).Return(pgx.ErrNoRows)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetEpisodePlayback, int64(2)).Return(row)

		repo := NewMovieRepo(&Client{Pool: pool})
		_, err := repo.GetEpisodePlayback(context.Background(), 2)
		if !errors.Is(err, ErrEpisodeNotFound) {
			t.Fatalf("expected ErrEpisodeNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		expectPlaybackRowScan(row, moviedomain.EpisodePlaybackResponse{
			EpisodeID:       2,
			MovieID:         1,
			SeasonNumber:    1,
			EpisodeNumber:   2,
			Title:           "Episode 2",
			DurationSeconds: 3600,
			PlaybackURL:     "videos/2.mp4",
		})

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetEpisodePlayback, int64(2)).Return(row)

		repo := NewMovieRepo(&Client{Pool: pool})
		got, err := repo.GetEpisodePlayback(context.Background(), 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.PlaybackURL != "videos/2.mp4" {
			t.Fatalf("unexpected playback: %#v", got)
		}
	})
}

func TestMovieRepo_GetWatchProgress(t *testing.T) {
	t.Parallel()

	t.Run("no rows", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().Scan(anyArgs(1)...).Return(pgx.ErrNoRows)

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetWatchProgress, int64(1), int64(2)).Return(row)

		repo := NewMovieRepo(&Client{Pool: pool})
		got, err := repo.GetWatchProgress(context.Background(), 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != 0 {
			t.Fatalf("expected zero progress, got %d", got)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		row := NewMockRow(ctrl)
		row.EXPECT().
			Scan(anyArgs(1)...).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 77

				return nil
			})

		pool := NewMockPool(ctrl)
		pool.EXPECT().QueryRow(gomock.Any(), sqlGetWatchProgress, int64(1), int64(2)).Return(row)

		repo := NewMovieRepo(&Client{Pool: pool})
		got, err := repo.GetWatchProgress(context.Background(), 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != 77 {
			t.Fatalf("expected progress 77, got %d", got)
		}
	})
}

func TestMovieRepo_UpsertWatchProgress(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pool := NewMockPool(ctrl)
	pool.EXPECT().
		Exec(gomock.Any(), sqlUpsertWatchProgress, int64(1), int64(2), 99).
		Return(pgconn.NewCommandTag("INSERT 1"), nil)

	repo := NewMovieRepo(&Client{Pool: pool})
	if err := repo.UpsertWatchProgress(context.Background(), 1, 2, 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
