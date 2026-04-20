package postgres

import (
	"fmt"
	"reflect"

	moviedomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	userdomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"go.uber.org/mock/gomock"
)

func anyArgs(n int) []any {
	args := make([]any, n)
	for i := range args {
		args[i] = gomock.Any()
	}

	return args
}

func inOrderCalls(calls []*gomock.Call) {
	args := make([]any, len(calls))
	for i, call := range calls {
		args[i] = call
	}

	gomock.InOrder(args...)
}

func assignScanDest(dest []any, values ...any) error {
	if len(dest) != len(values) {
		return fmt.Errorf("scan dest/value length mismatch: got %d dest, want %d values", len(dest), len(values))
	}

	for i, value := range values {
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Ptr || target.IsNil() {
			return fmt.Errorf("scan dest at index %d is not a non-nil pointer", i)
		}

		valueRef := reflect.ValueOf(value)
		targetElem := target.Elem()

		if !valueRef.IsValid() {
			targetElem.SetZero()
			continue
		}

		if !valueRef.Type().AssignableTo(targetElem.Type()) {
			return fmt.Errorf(
				"scan dest at index %d has incompatible type: got %s, want %s",
				i,
				targetElem.Type(),
				valueRef.Type(),
			)
		}

		targetElem.Set(valueRef)
	}

	return nil
}

func expectUserRowScan(row *MockRow, user userdomain.User) {
	row.EXPECT().
		Scan(anyArgs(9)...).
		DoAndReturn(func(dest ...any) error {
			return assignScanDest(dest,
				user.ID,
				user.Email,
				user.Password,
				user.Birthdate,
				user.AvatarFileKey,
				user.RegistrationDate,
				user.IsActive,
				user.CreatedAt,
				user.UpdatedAt,
			)
		})
}

func expectMovieRowScan(row *MockRow, movie moviedomain.MovieResponse) {
	row.EXPECT().
		Scan(anyArgs(15)...).
		DoAndReturn(func(dest ...any) error {
			return assignScanDest(dest,
				movie.ID,
				movie.Title,
				movie.Description,
				movie.Director,
				movie.TrailerURL,
				movie.ContentType,
				movie.ReleaseYear,
				movie.DurationSeconds,
				movie.AgeLimit,
				movie.OriginalLanguageID,
				movie.OriginalLanguage,
				movie.CountryID,
				movie.Country,
				movie.PictureFileKey,
				movie.PosterFileKey,
			)
		})
}

func expectActorRowScan(row *MockRow, actor moviedomain.ActorResponse) {
	row.EXPECT().
		Scan(anyArgs(6)...).
		DoAndReturn(func(dest ...any) error {
			return assignScanDest(dest,
				actor.ID,
				actor.FullName,
				actor.BirthDate,
				actor.Biography,
				actor.CountryID,
				actor.PictureFileKey,
			)
		})
}

func expectPlaybackRowScan(row *MockRow, playback moviedomain.EpisodePlaybackResponse) {
	row.EXPECT().
		Scan(anyArgs(7)...).
		DoAndReturn(func(dest ...any) error {
			return assignScanDest(dest,
				playback.EpisodeID,
				playback.MovieID,
				playback.SeasonNumber,
				playback.EpisodeNumber,
				playback.Title,
				playback.DurationSeconds,
				playback.PlaybackURL,
			)
		})
}

func expectStringRows(rows *MockRows, values []string) {
	calls := make([]*gomock.Call, 0, len(values)*2+3)

		for _, value := range values {
			value := value
			calls = append(calls, rows.EXPECT().Next().Return(true))
			calls = append(calls, rows.EXPECT().Scan(anyArgs(1)...).DoAndReturn(func(dest ...any) error {
				return assignScanDest(dest, value)
			}))
		}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}

func expectSelectionMoviePreviewRows(rows *MockRows, previews []moviedomain.MoviePreview) {
	calls := make([]*gomock.Call, 0, len(previews)*2+2)

		for _, preview := range previews {
			preview := preview
			calls = append(calls, rows.EXPECT().Next().Return(true))
			calls = append(calls, rows.EXPECT().Scan(anyArgs(3)...).DoAndReturn(func(dest ...any) error {
				return assignScanDest(dest, preview.ID, preview.Title, preview.ImgUrl)
			}))
		}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}

func expectMoviePreviewRows(rows *MockRows, previews []moviedomain.MoviePreview) {
	calls := make([]*gomock.Call, 0, len(previews)*2+3)

		for _, preview := range previews {
			preview := preview
			calls = append(calls, rows.EXPECT().Next().Return(true))
			calls = append(calls, rows.EXPECT().Scan(anyArgs(3)...).DoAndReturn(func(dest ...any) error {
				return assignScanDest(dest, preview.ID, preview.Title, preview.ImgUrl)
			}))
		}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}

func expectActorPreviewRows(rows *MockRows, actors []moviedomain.ActorPreview) {
	calls := make([]*gomock.Call, 0, len(actors)*2+3)

		for _, actor := range actors {
			actor := actor
			calls = append(calls, rows.EXPECT().Next().Return(true))
			calls = append(calls, rows.EXPECT().Scan(anyArgs(3)...).DoAndReturn(func(dest ...any) error {
				return assignScanDest(dest, actor.ID, actor.FullName, actor.PictureFileKey)
			}))
		}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}

func expectEpisodeRows(rows *MockRows, episodes []moviedomain.EpisodeItemResponse) {
	calls := make([]*gomock.Call, 0, len(episodes)*2+3)

		for _, episode := range episodes {
			episode := episode
			calls = append(calls, rows.EXPECT().Next().Return(true))
			calls = append(calls, rows.EXPECT().Scan(anyArgs(9)...).DoAndReturn(func(dest ...any) error {
				return assignScanDest(
					dest,
					episode.ID,
					episode.MovieID,
					episode.SeasonNumber,
					episode.EpisodeNumber,
					episode.Title,
					episode.Description,
					episode.DurationSeconds,
					episode.ImgURL,
					episode.VideoURL,
				)
			}))
		}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}
