package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/grpcx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpx/respond"
	authmw "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
)

func requireAuth(w http.ResponseWriter, r *http.Request) (authmw.AuthContext, bool) {
	authCtx, err := authmw.AuthFromContext(r.Context())
	if err != nil {
		httppkg.ErrResponse(w, http.StatusUnauthorized, "unauthorized")
		return authmw.AuthContext{}, false
	}

	return authCtx, true
}

func grpcContext(r *http.Request, timeout time.Duration) (contextDone func()) {
	ctx, cancel := grpcx.WithTimeout(r.Context(), timeout)
	*r = *r.WithContext(ctx)
	return cancel
}

func parsePathID(w http.ResponseWriter, r *http.Request, name string, message string) (int64, bool) {
	value := r.PathValue(name)
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