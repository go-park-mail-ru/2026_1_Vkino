package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"go.uber.org/mock/gomock"
)

func TestSupportWSAuthorization(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/support/tickets/1/subscribe", nil)
	req.Header.Set("Authorization", "Bearer token")

	got := supportWSAuthorization(req)
	if got != "Bearer token" {
		t.Fatalf("supportWSAuthorization() = %q, want %q", got, "Bearer token")
	}
}

func TestSupportWSAuthorization_QueryToken(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/support/tickets/1/subscribe?access_token=abc", nil)
	got := supportWSAuthorization(req)

	if got != "Bearer abc" {
		t.Fatalf("supportWSAuthorization() = %q, want %q", got, "Bearer abc")
	}
}

func TestSupportWSAuthorization_Empty(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/support/tickets/1/subscribe", nil)
	got := supportWSAuthorization(req)

	if got != "" {
		t.Fatalf("supportWSAuthorization() = %q, want empty", got)
	}
}

func TestSupportTicketSubscribeHandler_InvalidID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newSupportTicketSubscribeHandler(client)
	server := httpserver.New(httpserver.WithRoute("GET /support/tickets/{id}/subscribe", handler))

	rr := doRequest(server.Handler(), http.MethodGet, "/support/tickets/abc/subscribe", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid ticket id")
}

func TestSupportTicketSubscribeHandler_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	handler := newSupportTicketSubscribeHandler(client)
	server := httpserver.New(httpserver.WithRoute("GET /support/tickets/{id}/subscribe", handler))

	rr := doRequest(server.Handler(), http.MethodGet, "/support/tickets/10/subscribe", nil)

	requireJSONError(t, rr, http.StatusUnauthorized, "unauthorized")
}
