package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	moviedomain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/usecase/mocks"
	userusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	usermiddleware "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	"go.uber.org/mock/gomock"
)

func decodeBody[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rr.Body.String())
	}

	return v
}

func assertJSONContainsStringValue(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rr.Body.String())
	}

	for _, value := range body {
		if s, ok := value.(string); ok && s == want {
			return
		}
	}

	t.Fatalf("expected body to contain %q, got %v", want, body)
}

func authContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), usermiddleware.AuthCtxKey, userusecase.AuthContext{
		UserId: userID,
		Email:  "user@example.com",
	})

	return req.WithContext(ctx)
}

func TestHandler_GetAllSelections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantTitle  string
		wantError  string
	}{
		{
			name: "usecase error",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetAllSelections(gomock.Any()).
					Return(nil, postgresrepo.ErrSelectionNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "selection not found",
		},
		{
			name: "success",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetAllSelections(gomock.Any()).
					Return([]moviedomain.SelectionResponse{{Title: "popular"}}, nil)
			},
			wantStatus: http.StatusOK,
			wantTitle:  "popular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			tt.setupMocks(mu)

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/movie/selection/all", nil)
			rr := httptest.NewRecorder()

			h.GetAllSelections(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[[]moviedomain.SelectionResponse](t, rr)
			if len(got) != 1 || got[0].Title != tt.wantTitle {
				t.Fatalf("expected selection title %q, got %#v", tt.wantTitle, got)
			}
		})
	}
}

func TestHandler_GetSelectionByTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantTitle  string
		wantError  string
	}{
		{
			name:       "empty title",
			path:       "/movie/selection/",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid selection",
		},
		{
			name: "usecase error",
			path: "/movie/selection/popular",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetSelectionByTitle(gomock.Any(), "popular").
					Return(moviedomain.SelectionResponse{}, postgresrepo.ErrSelectionNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "selection not found",
		},
		{
			name: "success",
			path: "/movie/selection/popular",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetSelectionByTitle(gomock.Any(), "popular").
					Return(moviedomain.SelectionResponse{Title: "popular"}, nil)
			},
			wantStatus: http.StatusOK,
			wantTitle:  "popular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			h.GetSelectionByTitle(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.SelectionResponse](t, rr)
			if got.Title != tt.wantTitle {
				t.Fatalf("expected title %q, got %q", tt.wantTitle, got.Title)
			}
		})
	}
}

func TestHandler_GetMovieByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantTitle  string
		wantError  string
	}{
		{
			name:       "missing id",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid movie id",
		},
		{
			name:       "invalid id",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid movie id",
		},
		{
			name: "usecase error",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetMovieByID(gomock.Any(), int64(3)).
					Return(moviedomain.MovieResponse{}, postgresrepo.ErrMovieNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "movie not found",
		},
		{
			name: "success",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetMovieByID(gomock.Any(), int64(3)).
					Return(moviedomain.MovieResponse{ID: 3, Title: "Dune"}, nil)
			},
			wantStatus: http.StatusOK,
			wantTitle:  "Dune",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/movie/"+tt.id, nil)
			if tt.id != "" {
				req.SetPathValue("id", tt.id)
			}
			rr := httptest.NewRecorder()

			h.GetMovieByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.MovieResponse](t, rr)
			if got.Title != tt.wantTitle {
				t.Fatalf("expected title %q, got %q", tt.wantTitle, got.Title)
			}
		})
	}
}

func TestHandler_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantMovies int
		wantActors int
		wantError  string
	}{
		{
			name:  "invalid query",
			query: "",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					Search(gomock.Any(), "").
					Return(moviedomain.SearchResponse{}, moviedomain.ErrInvalidSearchQuery)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid search query",
		},
		{
			name:  "success",
			query: "dune",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					Search(gomock.Any(), "dune").
					Return(moviedomain.SearchResponse{
						Query:  "dune",
						Movies: []moviedomain.MoviePreview{{ID: 1, Title: "Dune"}},
						Actors: []moviedomain.ActorPreview{{ID: 2, FullName: "Zendaya"}},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantMovies: 1,
			wantActors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/movie/search?query="+tt.query, nil)
			rr := httptest.NewRecorder()

			h.Search(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.SearchResponse](t, rr)
			if len(got.Movies) != tt.wantMovies || len(got.Actors) != tt.wantActors {
				t.Fatalf("unexpected search result: %#v", got)
			}
		})
	}
}

func TestHandler_GetActorByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantName   string
		wantError  string
	}{
		{
			name:       "missing id",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid actor id",
		},
		{
			name:       "invalid id",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid actor id",
		},
		{
			name: "usecase error",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetActorByID(gomock.Any(), int64(3)).
					Return(moviedomain.ActorResponse{}, postgresrepo.ErrActorNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "actor not found",
		},
		{
			name: "success",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetActorByID(gomock.Any(), int64(3)).
					Return(moviedomain.ActorResponse{ID: 3, FullName: "Actor"}, nil)
			},
			wantStatus: http.StatusOK,
			wantName:   "Actor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/movie/actor/"+tt.id, nil)
			if tt.id != "" {
				req.SetPathValue("id", tt.id)
			}
			rr := httptest.NewRecorder()

			h.GetActorByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.ActorResponse](t, rr)
			if got.FullName != tt.wantName {
				t.Fatalf("expected name %q, got %q", tt.wantName, got.FullName)
			}
		})
	}
}

func TestHandler_GetEpisodePlayback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantURL    string
		wantError  string
	}{
		{
			name:       "missing id",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid episode id",
		},
		{
			name:       "invalid id",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid episode id",
		},
		{
			name: "usecase error",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(3), int64(0)).
					Return(moviedomain.EpisodePlaybackResponse{}, postgresrepo.ErrEpisodeNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantError:  "episode not found",
		},
		{
			name: "success",
			id:   "3",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetEpisodePlayback(gomock.Any(), int64(3), int64(0)).
					Return(moviedomain.EpisodePlaybackResponse{EpisodeID: 3, PlaybackURL: "https://cdn.example/video.mp4"}, nil)
			},
			wantStatus: http.StatusOK,
			wantURL:    "https://cdn.example/video.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/episode/"+tt.id+"/playback", nil)
			if tt.id != "" {
				req.SetPathValue("id", tt.id)
			}
			rr := httptest.NewRecorder()

			h.GetEpisodePlayback(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.EpisodePlaybackResponse](t, rr)
			if got.PlaybackURL != tt.wantURL {
				t.Fatalf("expected playback url %q, got %q", tt.wantURL, got.PlaybackURL)
			}
		})
	}
}

func TestHandler_GetEpisodeProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		withAuth   bool
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantPos    int
		wantError  string
	}{
		{
			name:       "unauthorized",
			id:         "3",
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized",
		},
		{
			name:       "invalid id",
			id:         "abc",
			withAuth:   true,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid episode id",
		},
		{
			name:     "usecase error",
			id:       "3",
			withAuth: true,
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetEpisodeProgress(gomock.Any(), int64(7), int64(3)).
					Return(moviedomain.WatchProgressResponse{}, moviedomain.ErrInvalidEpisodeID)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid episode id",
		},
		{
			name:     "success",
			id:       "3",
			withAuth: true,
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					GetEpisodeProgress(gomock.Any(), int64(7), int64(3)).
					Return(moviedomain.WatchProgressResponse{EpisodeID: 3, PositionSeconds: 41}, nil)
			},
			wantStatus: http.StatusOK,
			wantPos:    41,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodGet, "/episode/"+tt.id+"/progress", nil)
			if tt.id != "" {
				req.SetPathValue("id", tt.id)
			}
			if tt.withAuth {
				req = authContext(req, 7)
			}
			rr := httptest.NewRecorder()

			h.GetEpisodeProgress(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.WatchProgressResponse](t, rr)
			if got.PositionSeconds != tt.wantPos {
				t.Fatalf("expected position %d, got %d", tt.wantPos, got.PositionSeconds)
			}
		})
	}
}

func TestHandler_SaveEpisodeProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		withAuth   bool
		body       string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantPos    int
		wantError  string
	}{
		{
			name:       "unauthorized",
			id:         "3",
			body:       `{"position_seconds":10}`,
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized",
		},
		{
			name:       "invalid id",
			id:         "abc",
			withAuth:   true,
			body:       `{"position_seconds":10}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid episode id",
		},
		{
			name:       "invalid json",
			id:         "3",
			withAuth:   true,
			body:       `{"position_seconds":`,
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
		{
			name:     "usecase error",
			id:       "3",
			withAuth: true,
			body:     `{"position_seconds":10}`,
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					SaveEpisodeProgress(gomock.Any(), int64(7), int64(3), 10).
					Return(moviedomain.ErrInvalidWatchProgress)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid watch progress",
		},
		{
			name:     "success",
			id:       "3",
			withAuth: true,
			body:     `{"position_seconds":10}`,
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					SaveEpisodeProgress(gomock.Any(), int64(7), int64(3), 10).
					Return(nil)
			},
			wantStatus: http.StatusOK,
			wantPos:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(http.MethodPut, "/episode/"+tt.id+"/progress", strings.NewReader(tt.body))
			if tt.id != "" {
				req.SetPathValue("id", tt.id)
			}
			if tt.withAuth {
				req = authContext(req, 7)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.SaveEpisodeProgress(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantError != "" {
				assertJSONContainsStringValue(t, rr, tt.wantError)

				return
			}

			got := decodeBody[moviedomain.WatchProgressResponse](t, rr)
			if got.PositionSeconds != tt.wantPos {
				t.Fatalf("expected position %d, got %d", tt.wantPos, got.PositionSeconds)
			}
		})
	}
}
