package routes

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestPartyRoutes_JoinRoomByInviteBody(t *testing.T) {
	t.Parallel()

	client := &partyClientStub{
		joinResp: &partyv1.JoinRoomResponse{Room: &partyv1.Room{Id: 5}},
	}

	server := httpserver.New(Party(testConfig{}, client, nil, client)...)
	rr := doRequest(server.Handler(), http.MethodPost, "/watch-party/join", bytes.NewReader([]byte(`{"invite_link":"invite-123"}`)))

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, client.joinReq)
	require.Equal(t, "invite-123", client.joinReq.GetInviteLink())
}

func TestPartyRoutes_JoinRoomByInvitePath(t *testing.T) {
	t.Parallel()

	client := &partyClientStub{
		joinResp: &partyv1.JoinRoomResponse{Room: &partyv1.Room{Id: 5}},
	}

	server := httpserver.New(Party(testConfig{}, client, nil, client)...)
	rr := doRequest(server.Handler(), http.MethodGet, "/watch-party/join/invite-123", nil)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, client.joinReq)
	require.Equal(t, "invite-123", client.joinReq.GetInviteLink())
}

func TestPartyRoutes_InviteFriendToRoom(t *testing.T) {
	t.Parallel()

	client := &partyClientStub{
		friendResp: &userv1.GetFriendResponse{
			Id:        9,
			Email:     "friend@example.com",
			AvatarUrl: "https://cdn.example.com/avatar.png",
		},
		inviteFriendResp: &partyv1.InviteFriendToRoomResponse{
			RoomId:        5,
			InvitedUserId: 9,
			Status:        "pending",
		},
	}

	server := httpserver.New(Party(testConfig{}, client, nil, client)...)
	rr := doRequest(server.Handler(), http.MethodPost, "/watch-party/friends/9/invite", bytes.NewReader([]byte(`{"room_id":5}`)))

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, client.friendReq)
	require.NotNil(t, client.inviteFriendReq)
	require.Equal(t, int64(9), client.friendReq.GetFriendId())
	require.Equal(t, int64(5), client.inviteFriendReq.GetRoomId())
	require.Equal(t, int64(9), client.inviteFriendReq.GetInvitedUserId())
}

type partyClientStub struct {
	friendReq        *userv1.GetFriendRequest
	friendResp       *userv1.GetFriendResponse
	inviteReq        *partyv1.GetRoomInviteRequest
	inviteResp       *partyv1.GetRoomInviteResponse
	inviteFriendReq  *partyv1.InviteFriendToRoomRequest
	inviteFriendResp *partyv1.InviteFriendToRoomResponse
	joinReq          *partyv1.JoinRoomRequest
	joinResp         *partyv1.JoinRoomResponse
}

func (s *partyClientStub) GetOverview(context.Context, *partyv1.GetOverviewRequest, ...grpc.CallOption) (*partyv1.GetOverviewResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) GetRoom(context.Context, *partyv1.GetRoomRequest, ...grpc.CallOption) (*partyv1.GetRoomResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) GetRoomInvite(_ context.Context, in *partyv1.GetRoomInviteRequest, _ ...grpc.CallOption) (*partyv1.GetRoomInviteResponse, error) {
	s.inviteReq = in

	return s.inviteResp, nil
}

func (s *partyClientStub) InviteFriendToRoom(_ context.Context, in *partyv1.InviteFriendToRoomRequest, _ ...grpc.CallOption) (*partyv1.InviteFriendToRoomResponse, error) {
	s.inviteFriendReq = in

	return s.inviteFriendResp, nil
}

func (s *partyClientStub) CreateRoom(context.Context, *partyv1.CreateRoomRequest, ...grpc.CallOption) (*partyv1.CreateRoomResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) GetFriend(_ context.Context, in *userv1.GetFriendRequest, _ ...grpc.CallOption) (*userv1.GetFriendResponse, error) {
	s.friendReq = in

	return s.friendResp, nil
}

func (s *partyClientStub) JoinRoom(_ context.Context, in *partyv1.JoinRoomRequest, _ ...grpc.CallOption) (*partyv1.JoinRoomResponse, error) {
	s.joinReq = in

	return s.joinResp, nil
}

func (s *partyClientStub) DeleteRoom(context.Context, *partyv1.DeleteRoomRequest, ...grpc.CallOption) (*partyv1.DeleteRoomResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) ApplyRoomAction(context.Context, *partyv1.ApplyRoomActionRequest, ...grpc.CallOption) (*partyv1.ApplyRoomActionResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) SendRoomMessage(context.Context, *partyv1.SendRoomMessageRequest, ...grpc.CallOption) (*partyv1.SendRoomMessageResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) CreateRoomPoll(context.Context, *partyv1.CreateRoomPollRequest, ...grpc.CallOption) (*partyv1.CreateRoomPollResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) VoteRoomPoll(context.Context, *partyv1.VoteRoomPollRequest, ...grpc.CallOption) (*partyv1.VoteRoomPollResponse, error) {
	panic("unexpected call")
}

func (s *partyClientStub) SubscribeRoom(context.Context, *partyv1.SubscribeRoomRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[partyv1.RoomEvent], error) {
	panic("unexpected call")
}
