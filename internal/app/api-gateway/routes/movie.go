package routes

import (
	"net/http"
	"strings"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

func Movie(
	cfg Config,
	movieClient moviev1.MovieServiceClient,
) []httpserver.Option {
	return []httpserver.Option{
		route("GET /movie/genres", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetAllGenres(r.Context(), &moviev1.GetAllGenresRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/selection/all", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetAllSelections(r.Context(), &moviev1.GetAllSelectionsRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/selection/{selection}", func(w http.ResponseWriter, r *http.Request) {
			title := strings.TrimSpace(r.PathValue("selection"))
			if title == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid selection title")

				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetSelectionByTitle(r.Context(), &moviev1.GetSelectionByTitleRequest{
				Title: title,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/search", func(w http.ResponseWriter, r *http.Request) {
			query := strings.TrimSpace(r.URL.Query().Get("query"))
			if query == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid search query")

				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.SearchMovies(r.Context(), &moviev1.SearchMoviesRequest{
				Query: query,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/genre/{id}", func(w http.ResponseWriter, r *http.Request) {
			genreID, ok := parsePathID(w, r, "invalid genre id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetGenreByID(r.Context(), &moviev1.GetGenreByIDRequest{
				GenreId: genreID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/{id}", func(w http.ResponseWriter, r *http.Request) {
			movieID, ok := parsePathID(w, r, "invalid movie id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetMovieByID(r.Context(), &moviev1.GetMovieByIDRequest{
				MovieId: movieID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /movie/actor/{id}", func(w http.ResponseWriter, r *http.Request) {
			actorID, ok := parsePathID(w, r, "invalid actor id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetActorByID(r.Context(), &moviev1.GetActorByIDRequest{
				ActorId: actorID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /episode/{id}/playback", func(w http.ResponseWriter, r *http.Request) {
			episodeID, ok := parsePathID(w, r, "invalid episode id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetEpisodePlayback(r.Context(), &moviev1.GetEpisodePlaybackRequest{
				EpisodeId: episodeID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("GET /episode/{id}/progress", func(w http.ResponseWriter, r *http.Request) {
			episodeID, ok := parsePathID(w, r, "invalid episode id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.GetEpisodeProgress(r.Context(), &moviev1.GetEpisodeProgressRequest{
				EpisodeId: episodeID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		route("PUT /episode/{id}/progress", func(w http.ResponseWriter, r *http.Request) {
			episodeID, ok := parsePathID(w, r, "invalid episode id")
			if !ok {
				return
			}

			var req struct {
				PositionSeconds int64 `json:"position_seconds"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.MovieRequestTimeout())
			defer cancel()

			resp, err := movieClient.SaveEpisodeProgress(r.Context(), &moviev1.SaveEpisodeProgressRequest{
				EpisodeId:       episodeID,
				PositionSeconds: req.PositionSeconds,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),
	}
}
