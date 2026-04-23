package routes

import (
	"errors"
	"io"
	"net/http"
	"strings"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
)

type updateProfileRequest struct {
	Birthdate string `json:"birthdate"`
}

type updateProfileJSONRequest struct {
	Birthdate string `json:"birthdate"`
}

type updateProfilePayload struct {
	Birthdate         string
	Avatar            []byte
	AvatarContentType string
}

func readUpdateProfilePayload(w http.ResponseWriter, r *http.Request) (updateProfilePayload, bool) {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		// лимит тела запроса, чтобы не тащить бесконечный файл в память
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid multipart form body")
			return updateProfilePayload{}, false
		}

		payload := updateProfilePayload{
			Birthdate: strings.TrimSpace(r.FormValue("birthdate")),
		}

		file, header, err := r.FormFile("avatar")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				return payload, true
			}

			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid avatar file")
			return updateProfilePayload{}, false
		}
		defer file.Close()

		avatarBytes, err := io.ReadAll(file)
		if err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "failed to read avatar file")
			return updateProfilePayload{}, false
		}

		payload.Avatar = avatarBytes
		if header != nil {
			payload.AvatarContentType = header.Header.Get("Content-Type")
		}

		return payload, true

	default:
		var req updateProfileJSONRequest
		if !readJSON(w, r, &req) {
			return updateProfilePayload{}, false
		}

		return updateProfilePayload{
			Birthdate: strings.TrimSpace(req.Birthdate),
		}, true
	}
}

func User(
	cfg Config,
	userClient userv1.UserServiceClient,
) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("GET /user/me", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.GetProfile(r.Context(), &userv1.GetProfileRequest{})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /user/search", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.SearchUsersByEmail(r.Context(), &userv1.SearchUsersByEmailRequest{
				EmailQuery: r.URL.Query().Get("email"),
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("PUT /user/profile", func(w http.ResponseWriter, r *http.Request) {
			req, ok := readUpdateProfilePayload(w, r)
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.UpdateProfile(r.Context(), &userv1.UpdateProfileRequest{
				Birthdate:         req.Birthdate,
				Avatar:            req.Avatar,
				AvatarContentType: req.AvatarContentType,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("POST /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			friendID, ok := parsePathID(w, r, "id", "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.AddFriend(r.Context(), &userv1.AddFriendRequest{
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("DELETE /user/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
			friendID, ok := parsePathID(w, r, "id", "invalid friend id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			_, err := userClient.DeleteFriend(r.Context(), &userv1.DeleteFriendRequest{
				FriendId: friendID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, map[string]bool{
				"success": true,
			})
		}),

		httpserver.WithRoute("PUT /user/favorites/{id}", func(w http.ResponseWriter, r *http.Request) {
			movieID, ok := parsePathID(w, r, "id", "invalid movie id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := userClient.AddMovieToFavorites(r.Context(), &userv1.AddMovieToFavoritesRequest{
				MovieId: movieID,
			})
			if err != nil {
				writeGRPCError(w, err)
				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),
	}
}
