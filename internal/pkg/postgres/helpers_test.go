package postgres

import (
	"time"

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

func expectUserRowScan(row *MockRow, user userdomain.User) {
	row.EXPECT().
		Scan(anyArgs(9)...).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = user.ID
			*dest[1].(*string) = user.Email
			*dest[2].(*string) = user.Password
			*dest[3].(**time.Time) = user.Birthdate
			*dest[4].(**string) = user.AvatarFileKey
			*dest[5].(*time.Time) = user.RegistrationDate
			*dest[6].(*bool) = user.IsActive
			*dest[7].(*time.Time) = user.CreatedAt
			*dest[8].(*time.Time) = user.UpdatedAt

			return nil
		})
}

func expectMovieRowScan(row *MockRow, movie moviedomain.MovieResponse) {
	row.EXPECT().
		Scan(anyArgs(11)...).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = movie.ID
			*dest[1].(*string) = movie.Title
			*dest[2].(*string) = movie.Description
			*dest[3].(*string) = movie.Director
			*dest[4].(*string) = movie.ContentType
			*dest[5].(*int) = movie.ReleaseYear
			*dest[6].(*int) = movie.DurationSeconds
			*dest[7].(*int) = movie.AgeLimit
			*dest[8].(*int64) = movie.OriginalLanguageID
			*dest[9].(*int64) = movie.CountryID
			*dest[10].(*string) = movie.PictureFileKey

			return nil
		})
}

func expectActorRowScan(row *MockRow, actor moviedomain.ActorResponse) {
	row.EXPECT().
		Scan(anyArgs(6)...).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = actor.ID
			*dest[1].(*string) = actor.FullName
			*dest[2].(*string) = actor.BirthDate
			*dest[3].(*string) = actor.Biography
			*dest[4].(*int64) = actor.CountryID
			*dest[5].(*string) = actor.PictureFileKey

			return nil
		})
}

func expectPlaybackRowScan(row *MockRow, playback moviedomain.EpisodePlaybackResponse) {
	row.EXPECT().
		Scan(anyArgs(7)...).
		DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = playback.EpisodeID
			*dest[1].(*int64) = playback.MovieID
			*dest[2].(*int) = playback.SeasonNumber
			*dest[3].(*int) = playback.EpisodeNumber
			*dest[4].(*string) = playback.Title
			*dest[5].(*int) = playback.DurationSeconds
			*dest[6].(*string) = playback.PlaybackURL

			return nil
		})
}

func expectStringRows(rows *MockRows, values []string) {
	calls := make([]*gomock.Call, 0, len(values)*2+3)

	for _, value := range values {
		value := value
		calls = append(calls, rows.EXPECT().Next().Return(true))
		calls = append(calls, rows.EXPECT().Scan(anyArgs(1)...).DoAndReturn(func(dest ...any) error {
			*dest[0].(*string) = value

			return nil
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
			*dest[0].(*int64) = preview.ID
			*dest[1].(*string) = preview.Title
			*dest[2].(*string) = preview.ImgUrl

			return nil
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
			*dest[0].(*int64) = preview.ID
			*dest[1].(*string) = preview.Title
			*dest[2].(*string) = preview.ImgUrl

			return nil
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
			*dest[0].(*int64) = actor.ID
			*dest[1].(*string) = actor.FullName
			*dest[2].(*string) = actor.PictureFileKey

			return nil
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
		calls = append(calls, rows.EXPECT().Scan(anyArgs(8)...).DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = episode.ID
			*dest[1].(*int64) = episode.MovieID
			*dest[2].(*int) = episode.SeasonNumber
			*dest[3].(*int) = episode.EpisodeNumber
			*dest[4].(*string) = episode.Title
			*dest[5].(*string) = episode.Description
			*dest[6].(*int) = episode.DurationSeconds
			*dest[7].(*string) = episode.ImgURL

			return nil
		}))
	}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}

func expectUserSearchRows(rows *MockRows, users []userdomain.UserSearchResult) {
	calls := make([]*gomock.Call, 0, len(users)*2+3)

	for _, user := range users {
		user := user
		calls = append(calls, rows.EXPECT().Next().Return(true))
		calls = append(calls, rows.EXPECT().Scan(anyArgs(3)...).DoAndReturn(func(dest ...any) error {
			*dest[0].(*int64) = user.ID
			*dest[1].(*string) = user.Email
			*dest[2].(*bool) = user.IsFriend

			return nil
		}))
	}

	calls = append(calls, rows.EXPECT().Next().Return(false))
	calls = append(calls, rows.EXPECT().Err().Return(nil))
	calls = append(calls, rows.EXPECT().Close())

	inOrderCalls(calls)
}
