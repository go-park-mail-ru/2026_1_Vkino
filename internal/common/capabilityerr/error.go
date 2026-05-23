package capabilityerr

import (
	"encoding/json"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errorDomain = "subscription.vkino"

type Detail struct {
	Code          string
	Feature       string
	Message       string
	CurrentLevel  *int32
	RequiredLevel *int32
	Limit         *int32
	Used          *int32
	Remaining     *int32
}

type Error struct {
	detail Detail
}

type grpcDetailPayload struct {
	Domain        string `json:"domain"`
	Code          string `json:"code"`
	Feature       string `json:"feature"`
	Message       string `json:"message"`
	CurrentLevel  *int32 `json:"current_level,omitempty"`
	RequiredLevel *int32 `json:"required_level,omitempty"`
	Limit         *int32 `json:"limit,omitempty"`
	Used          *int32 `json:"used,omitempty"`
	Remaining     *int32 `json:"remaining,omitempty"`
}

func New(detail Detail) error {
	return &Error{detail: detail}
}

func (e *Error) Error() string {
	return e.detail.Message
}

func (e *Error) Detail() Detail {
	return e.detail
}

func ToGRPCError(err error) (error, bool) {
	var subscriptionErr *Error
	if !errors.As(err, &subscriptionErr) {
		return nil, false
	}

	st := status.New(codes.PermissionDenied, encodeGRPCDetail(subscriptionErr.Detail()))

	return st.Err(), true
}

func DetailFromGRPCError(err error) (Detail, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return Detail{}, false
	}

	var payload grpcDetailPayload
	if err := json.Unmarshal([]byte(st.Message()), &payload); err != nil {
		return Detail{}, false
	}

	if payload.Domain != errorDomain {
		return Detail{}, false
	}

	return Detail{
		Code:          payload.Code,
		Feature:       payload.Feature,
		Message:       payload.Message,
		CurrentLevel:  payload.CurrentLevel,
		RequiredLevel: payload.RequiredLevel,
		Limit:         payload.Limit,
		Used:          payload.Used,
		Remaining:     payload.Remaining,
	}, true
}

func encodeGRPCDetail(detail Detail) string {
	payload := grpcDetailPayload{
		Domain:        errorDomain,
		Code:          detail.Code,
		Feature:       detail.Feature,
		Message:       detail.Message,
		CurrentLevel:  detail.CurrentLevel,
		RequiredLevel: detail.RequiredLevel,
		Limit:         detail.Limit,
		Used:          detail.Used,
		Remaining:     detail.Remaining,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return detail.Message
	}

	return string(encoded)
}
