//nolint:wsl_v5 // Tiny route bootstrap stays clearer in this compact form.
package routes

import (
	"net/http"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

func Register(
	cfg Config,
	authClient authv1.AuthServiceClient,
	userClient UserClient,
	movieClient moviev1.MovieServiceClient,
) []httpserver.Option {
	result := make([]httpserver.Option, 0, 1+len(Auth(cfg, authClient))+len(User(cfg, userClient))+
		len(Movie(cfg, movieClient)))
	result = append(result,
		route("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok")); err != nil {
				return
			}
		}),
	)

	result = append(result, Auth(cfg, authClient)...)
	result = append(result, User(cfg, userClient)...)
	result = append(result, Movie(cfg, movieClient)...)

	return result
}
