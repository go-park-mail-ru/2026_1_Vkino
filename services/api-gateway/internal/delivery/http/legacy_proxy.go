package http

import "net/http"

type LegacyProxyHandler struct {
	proxy http.Handler
}

func NewLegacyProxyHandler(proxy http.Handler) *LegacyProxyHandler {
	return &LegacyProxyHandler{proxy: proxy}
}

func (h *LegacyProxyHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	h.proxy.ServeHTTP(w, r)
}
