package gateway

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	deliveryhttp "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http"
)

func registerRoutes(
	healthHandler *deliveryhttp.HealthHandler,
	authHandler *deliveryhttp.AuthHandler,
	legacyHandler *deliveryhttp.LegacyProxyHandler,
	authMiddleware AuthMiddleware,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("GET /healthz", healthHandler.Health),

		httpserver.WithRoute("POST /user/sign-up", authHandler.SignUp),
		httpserver.WithRoute("POST /user/sign-in", authHandler.SignIn),
		httpserver.WithRoute("POST /user/refresh", authHandler.Refresh),

		httpserver.WithMiddlewareRoute("POST /user/logout", authHandler.LogOut, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/change-password", authHandler.ChangePassword, authMiddleware.Middleware),

		httpserver.WithRoute("GET /user/me", legacyHandler.Proxy),
		httpserver.WithRoute("PUT /user/profile", legacyHandler.Proxy),

		httpserver.WithRoute("GET /movie/selection/all", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/selection/{selection}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/{id}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/actor/{id}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /episode/{id}/playback", legacyHandler.Proxy),
		httpserver.WithRoute("GET /episode/{id}/progress", legacyHandler.Proxy),
		httpserver.WithRoute("PUT /episode/{id}/progress", legacyHandler.Proxy),
	}
}
