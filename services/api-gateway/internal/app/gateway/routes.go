package gateway

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	deliveryhttp "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http"
)

func registerRoutes(
	healthHandler *deliveryhttp.HealthHandler,
	authHandler *deliveryhttp.AuthHandler,
	userHandler *deliveryhttp.UserHandler,
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

		httpserver.WithMiddlewareRoute("GET /user/me", userHandler.GetProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("PUT /user/profile", userHandler.UpdateProfile, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("GET /user/search", userHandler.SearchUsersByEmail, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /user/friend", userHandler.AddFriend, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("DELETE /user/friend", userHandler.DeleteFriend, authMiddleware.Middleware),
		httpserver.WithMiddlewareRoute("POST /movie/{id}/favorite", userHandler.AddMovieToFavorites, authMiddleware.Middleware),

		httpserver.WithRoute("GET /movie/selection/all", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/selection/{selection}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/{id}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /movie/actor/{id}", legacyHandler.Proxy),
		httpserver.WithRoute("GET /episode/{id}/playback", legacyHandler.Proxy),
		httpserver.WithRoute("GET /episode/{id}/progress", legacyHandler.Proxy),
		httpserver.WithRoute("PUT /episode/{id}/progress", legacyHandler.Proxy),
	}
}