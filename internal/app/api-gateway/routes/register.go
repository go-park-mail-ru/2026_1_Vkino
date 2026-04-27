package routes

import (
	"net/http"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

func Register(cfg Config, authClient authv1.AuthServiceClient, userClient UserClient,
	supportFileStore storage.FileStorage, movieClient moviev1.MovieServiceClient) []httpserver.Option {
	result := make([]httpserver.Option, 0, 1+len(Auth(cfg, authClient))+len(User(cfg, userClient, supportFileStore))+
		len(Movie(cfg, movieClient)))
	result = append(result,
		httpserver.WithRoute("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	result = append(result, Auth(cfg, authClient)...)
	result = append(result, User(cfg, userClient, supportFileStore)...)
	result = append(result, Movie(cfg, movieClient)...)

	return result
}
