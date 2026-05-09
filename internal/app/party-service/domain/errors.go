package domain

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrAlreadyMember     = errors.New("already member")
	ErrRoomClosed        = errors.New("room closed")
	ErrInvalidRoomID     = errors.New("invalid room id")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidRoomName   = errors.New("invalid room name")
	ErrInvalidVisibility = errors.New("invalid visibility")
	ErrInvalidInviteLink = errors.New("invalid invite link")
	ErrInvalidEventType  = errors.New("invalid room event type")
	ErrInvalidPlayback   = errors.New("invalid playback state")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrNotImplemented    = errors.New("not implemented")
	ErrInternal          = errors.New("internal error")
)
