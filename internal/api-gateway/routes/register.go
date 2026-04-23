package routes

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

func Register(
	cfg Config,
	authClient authv1.AuthServiceClient,
	userClient userv1.UserServiceClient,
	movieClient moviev1.MovieServiceClient,
) []httpserver.Option {
	result := []httpserver.Option{
		httpserver.WithRoute("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}

	result = append(result, Auth(cfg, authClient)...)
	result = append(result, User(cfg, userClient)...)
	result = append(result, Movie(cfg, movieClient)...)

	return result
}
