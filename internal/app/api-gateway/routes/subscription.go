package routes

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/common/capabilityerr"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type subscriptionHTTPErrorResponse struct {
	Error subscriptionHTTPError `json:"error"`
}

type subscriptionHTTPError struct {
	Code          string `json:"code"`
	Feature       string `json:"feature"`
	Message       string `json:"message"`
	CurrentLevel  *int32 `json:"current_level,omitempty"`
	RequiredLevel *int32 `json:"required_level,omitempty"`
	Limit         *int32 `json:"limit,omitempty"`
	Used          *int32 `json:"used,omitempty"`
	Remaining     *int32 `json:"remaining,omitempty"`
}

func writeSubscriptionGRPCError(w http.ResponseWriter, err error) bool {
	detail, ok := capabilityerr.DetailFromGRPCError(err)
	if !ok {
		return false
	}

	httppkg.Response(w, http.StatusForbidden, subscriptionHTTPErrorResponse{
		Error: subscriptionHTTPError{
			Code:          detail.Code,
			Feature:       detail.Feature,
			Message:       detail.Message,
			CurrentLevel:  detail.CurrentLevel,
			RequiredLevel: detail.RequiredLevel,
			Limit:         detail.Limit,
			Used:          detail.Used,
			Remaining:     detail.Remaining,
		},
	})

	return true
}
