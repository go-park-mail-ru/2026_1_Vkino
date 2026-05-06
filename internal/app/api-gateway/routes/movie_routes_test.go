package routes

import (
	"bytes"
	"net/http"
	"testing"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMovieRoutes_GetAllGenres(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetAllGenres(gomock.Any(), &moviev1.GetAllGenresRequest{}).
		Return(&moviev1.GetAllGenresResponse{Genres: []*moviev1.GenreShort{{Id: 1, Title: "Drama"}}}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/genres", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetAllSelections(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetAllSelections(gomock.Any(), &moviev1.GetAllSelectionsRequest{}).
		Return(&moviev1.GetAllSelectionsResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/selection/all", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetSelectionByTitle(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetSelectionByTitle(gomock.Any(), &moviev1.GetSelectionByTitleRequest{Title: "Top"}).
		Return(&moviev1.GetSelectionByTitleResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/selection/%20Top%20", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetSelectionByTitle_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/selection/%20%20", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid selection title")
}

func TestMovieRoutes_SearchMovies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().SearchMovies(gomock.Any(), &moviev1.SearchMoviesRequest{Query: "alien"}).
		Return(&moviev1.SearchMoviesResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/search?query=%20alien%20", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_SearchMovies_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/search", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid search query")
}

func TestMovieRoutes_GetGenreByID_Numeric(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetGenreByID(gomock.Any(), &moviev1.GetGenreByIDRequest{GenreId: 12}).
		Return(&moviev1.GetGenreByIDResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/genre/12", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetGenreByID_Title(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	gomock.InOrder(
		client.EXPECT().GetAllGenres(gomock.Any(), &moviev1.GetAllGenresRequest{}).
			Return(&moviev1.GetAllGenresResponse{Genres: []*moviev1.GenreShort{{Id: 7, Title: "Comedy"}}}, nil),
		client.EXPECT().GetGenreByID(gomock.Any(), &moviev1.GetGenreByIDRequest{GenreId: 7}).
			Return(&moviev1.GetGenreByIDResponse{}, nil),
	)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/genre/Comedy", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetGenreByID_UnknownTitle(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetAllGenres(gomock.Any(), &moviev1.GetAllGenresRequest{}).
		Return(&moviev1.GetAllGenresResponse{Genres: []*moviev1.GenreShort{{Id: 1, Title: "Drama"}}}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/genre/Action", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid genre id")
}

func TestMovieRoutes_GetMovieByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetMovieByID(gomock.Any(), &moviev1.GetMovieByIDRequest{MovieId: 42}).
		Return(&moviev1.GetMovieByIDResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/42", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetMovieByID_Invalid(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/abc", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid movie id")
}

func TestMovieRoutes_GetMovieByID_GRPCError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetMovieByID(gomock.Any(), &moviev1.GetMovieByIDRequest{MovieId: 99}).
		Return(nil, status.Error(codes.NotFound, "movie not found"))

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/99", nil)

	requireJSONError(t, rr, http.StatusNotFound, "movie not found")
}

func TestMovieRoutes_GetActorByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetActorByID(gomock.Any(), &moviev1.GetActorByIDRequest{ActorId: 5}).
		Return(&moviev1.GetActorByIDResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/movie/actor/5", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetEpisodePlayback(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetEpisodePlayback(gomock.Any(), &moviev1.GetEpisodePlaybackRequest{EpisodeId: 3}).
		Return(&moviev1.GetEpisodePlaybackResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/episode/3/playback", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_GetEpisodeProgress(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().GetEpisodeProgress(gomock.Any(), &moviev1.GetEpisodeProgressRequest{EpisodeId: 11}).
		Return(&moviev1.GetEpisodeProgressResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/episode/11/progress", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_SaveEpisodeProgress(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	client.EXPECT().SaveEpisodeProgress(gomock.Any(), &moviev1.SaveEpisodeProgressRequest{
		EpisodeId:       7,
		PositionSeconds: 120,
	}).Return(&moviev1.SaveEpisodeProgressResponse{}, nil)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodPut, "/episode/7/progress", bytes.NewReader([]byte(`{"position_seconds":120}`)))

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestMovieRoutes_SaveEpisodeProgress_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockMovieServiceClient(ctrl)

	handler := newMovieHandler(t, client)
	rr := doRequest(handler, http.MethodPut, "/episode/7/progress", bytes.NewReader([]byte(`{"position_seconds":"bad"}`)))

	requireJSONError(t, rr, http.StatusBadRequest, "invalid json body")
}
