package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	"google.golang.org/grpc/codes"
)

var partyGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrRoomNotFound,
		domain.ErrInviteNotFound,
		domain.ErrAccessDenied,
		domain.ErrAlreadyMember,
		domain.ErrRoomClosed,
		domain.ErrInvalidRoomID,
		domain.ErrInvalidUserID,
		domain.ErrInvalidRoomName,
		domain.ErrInvalidVisibility,
		domain.ErrInvalidInviteLink,
		domain.ErrInvalidEventType,
		domain.ErrInvalidPlayback,
		domain.ErrInvalidMessage,
		domain.ErrNotImplemented,
		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrRoomNotFound:      {Code: codes.NotFound, Message: "room not found"},
		domain.ErrInviteNotFound:    {Code: codes.NotFound, Message: "invite not found"},
		domain.ErrAccessDenied:      {Code: codes.PermissionDenied, Message: "access denied"},
		domain.ErrAlreadyMember:     {Code: codes.AlreadyExists, Message: "already member"},
		domain.ErrRoomClosed:        {Code: codes.FailedPrecondition, Message: "room closed"},
		domain.ErrInvalidRoomID:     {Code: codes.InvalidArgument, Message: "invalid room id"},
		domain.ErrInvalidUserID:     {Code: codes.InvalidArgument, Message: "invalid user id"},
		domain.ErrInvalidRoomName:   {Code: codes.InvalidArgument, Message: "invalid room name"},
		domain.ErrInvalidVisibility: {Code: codes.InvalidArgument, Message: "invalid visibility"},
		domain.ErrInvalidInviteLink: {Code: codes.InvalidArgument, Message: "invalid invite link"},
		domain.ErrInvalidEventType:  {Code: codes.InvalidArgument, Message: "invalid room event type"},
		domain.ErrInvalidPlayback:   {Code: codes.InvalidArgument, Message: "invalid playback state"},
		domain.ErrInvalidMessage:    {Code: codes.InvalidArgument, Message: "invalid message"},
		domain.ErrNotImplemented:    {Code: codes.Unimplemented, Message: "not implemented"},
		domain.ErrInternal:          {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapError(err error) error {
	return partyGRPCErrorMapper.Map(err)
}
