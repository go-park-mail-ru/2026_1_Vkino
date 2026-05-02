package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/respond"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

const gatewayServiceName = "api-gateway"

func grpcContext(r *http.Request, timeout time.Duration) (contextDone func()) {
	ctx, cancel := grpcx.WithTimeout(r.Context(), timeout)

	if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
		ctx = authctx.AppendOutgoing(ctx, authorization)
	}

	*r = *r.WithContext(ctx)

	return cancel
}

func route(pattern string, handler http.HandlerFunc) httpserver.Option {
	return httpserver.WithRoute(
		pattern,
		metrics.InstrumentHTTPHandlerFunc(gatewayServiceName, routeLabel(pattern), handler),
	)
}

func routeLabel(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	_, route, found := strings.Cut(pattern, " ")
	if !found {
		return pattern
	}

	route = strings.TrimSpace(route)
	if route == "" {
		return pattern
	}

	return route
}

func parsePathID(w http.ResponseWriter, r *http.Request, message string) (int64, bool) {
	value := r.PathValue("id")

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	return id, true
}

func readJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) bool {
	if err := httppkg.Read(r, dst); err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, "invalid json body")

		return false
	}

	return true
}

func writeGRPCError(w http.ResponseWriter, err error) {
	respond.Error(w, err)
}
