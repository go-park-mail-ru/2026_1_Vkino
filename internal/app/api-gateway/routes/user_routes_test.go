package routes

import (
	"bytes"
	"net/http"
	"testing"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserRoutes_GetProfile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetProfile(gomock.Any(), &userv1.GetProfileRequest{}).
		Return(&userv1.GetProfileResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/me", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_SearchUsers_DefaultLimit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().SearchUsers(gomock.Any(), &userv1.SearchUsersRequest{Query: "alex", Limit: defaultSearchLimit}).
		Return(&userv1.SearchUsersResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/search?query=alex&limit=bad", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_UpdateProfile_JSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().UpdateProfile(gomock.Any(), &userv1.UpdateProfileRequest{Birthdate: "2000-01-01"}).
		Return(&userv1.UpdateProfileResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPut, "/user/profile", bytes.NewReader([]byte(`{"birthdate":"2000-01-01"}`)))

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_SendFriendRequest_InvalidID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/user/friends/abc", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid friend id")
}

func TestUserRoutes_DeleteFriend(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().DeleteFriend(gomock.Any(), &userv1.DeleteFriendRequest{FriendId: 9}).
		Return(&userv1.DeleteFriendResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodDelete, "/user/friends/9", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_ToggleFavorite(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().ToggleFavorite(gomock.Any(), &userv1.ToggleFavoriteRequest{MovieId: 4}).
		Return(&userv1.ToggleFavoriteResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPut, "/user/favorites/4", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetFavorites_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetFavorites(gomock.Any(), &userv1.GetFavoritesRequest{Limit: defaultSearchLimit, Offset: 0}).
		Return(&userv1.GetFavoritesResponse{MovieIds: nil, TotalCount: 0}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/favorites", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetFavorites_WithMovies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetFavorites(gomock.Any(), &userv1.GetFavoritesRequest{Limit: defaultSearchLimit, Offset: 0}).
		Return(&userv1.GetFavoritesResponse{MovieIds: []int64{2, 1}, TotalCount: 2}, nil)
	client.EXPECT().GetMoviesByIDs(gomock.Any(), &moviev1.GetMoviesByIDsRequest{MovieIds: []int64{2, 1}}).
		Return(&moviev1.GetMoviesByIDsResponse{Movies: []*moviev1.MovieCard{{Id: 1}, {Id: 2}}}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/favorites", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetContinueWatching(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetContinueWatching(gomock.Any(), &moviev1.GetContinueWatchingRequest{Limit: defaultContinueWatchingLimit}).
		Return(&moviev1.GetContinueWatchingResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/watch/continue", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetWatchHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetWatchHistory(gomock.Any(), &moviev1.GetWatchHistoryRequest{Limit: defaultSearchLimit, MinProgress: 0}).
		Return(&moviev1.GetWatchHistoryResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/watch/history", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetWatchRecent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetWatchHistory(gomock.Any(), &moviev1.GetWatchHistoryRequest{Limit: defaultSearchLimit, MinProgress: recentWatchHistoryMinProgress}).
		Return(&moviev1.GetWatchHistoryResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/watch/recent", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetFriendRequests(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetFriendRequests(gomock.Any(), &userv1.GetFriendRequestsRequest{Direction: "out", Limit: defaultCollectionLimit}).
		Return(&userv1.GetFriendRequestsResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/friends/requests?direction=out", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_RespondToFriendRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/user/friends/requests/10/respond", bytes.NewReader([]byte(`{"action":1}`)))

	requireJSONError(t, rr, http.StatusBadRequest, "invalid json body")
}

func TestUserRoutes_DeleteOutgoingFriendRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().DeleteOutgoingFriendRequest(gomock.Any(), &userv1.DeleteOutgoingFriendRequestRequest{RequestId: 11}).
		Return(&userv1.DeleteOutgoingFriendRequestResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodDelete, "/user/friends/requests/11", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetFriendsList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetFriendsList(gomock.Any(), &userv1.GetFriendsListRequest{Limit: defaultCollectionLimit, Offset: 0}).
		Return(&userv1.GetFriendsListResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/user/friends", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_CreateTicket(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().CreateTicket(gomock.Any(), &supportv1.CreateTicketRequest{
		Category:          "billing",
		Title:             "Help",
		Description:       "Details",
		UserEmail:         "user@example.com",
		AttachmentFileKey: "file",
	}).Return(&supportv1.TicketResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/support/tickets", bytes.NewReader([]byte(`{"category":"billing","title":"Help","description":"Details","user_email":" user@example.com ","attachment_file_key":"file"}`)))

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestUserRoutes_GetTickets_InvalidSupportLine(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/tickets?support_line=bad", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid support line")
}

func TestUserRoutes_UpdateTicket(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().UpdateTicket(gomock.Any(), &supportv1.UpdateTicketRequest{
		TicketId:          5,
		Category:          "billing",
		Status:            "open",
		SupportLine:       2,
		Title:             "Help",
		UserEmail:         "user@example.com",
		Description:       "Details",
		AttachmentFileKey: "file",
		Rating:            4,
	}).Return(&supportv1.TicketResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPatch, "/support/tickets/5", bytes.NewReader([]byte(`{"category":"billing","status":"open","support_line":2,"title":"Help","user_email":"user@example.com","description":"Details","attachment_file_key":"file","rating":4}`)))

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_GetTicketMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetTicketMessages(gomock.Any(), &supportv1.GetTicketMessagesRequest{TicketId: 2}).
		Return(&supportv1.TicketMessagesResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/tickets/2/messages", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRoutes_CreateTicketMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().CreateTicketMessage(gomock.Any(), &supportv1.CreateTicketMessageRequest{TicketId: 3, Content: "hello", ContentFileKey: "key"}).
		Return(&supportv1.TicketMessageResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/support/tickets/3/messages", bytes.NewReader([]byte(`{"content":"hello","content_file_key":"key"}`)))

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestUserRoutes_GetTicketStatistics(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().GetTicketStatistics(gomock.Any(), &supportv1.GetTicketStatisticsRequest{}).
		Return(&supportv1.TicketStatisticsResponse{}, nil)

	handler := newUserHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodGet, "/support/statistics", nil)

	require.Equal(t, http.StatusOK, rr.Code)
}
