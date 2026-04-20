package legacyhttp

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(baseURL string) (http.Handler, error) {
	target, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse legacy api url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		r.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return proxy, nil
}
