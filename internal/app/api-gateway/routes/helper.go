package routes

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
)

func parseInt32Query(r *http.Request, key string, defaultValue int32) int32 {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(parsed)
}

func resolveGenreID(
	ctx context.Context,
	movieClient moviev1.MovieServiceClient,
	raw string,
) (int64, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, false, nil
	}

	if genreID, err := strconv.ParseInt(value, 10, 64); err == nil {
		return genreID, true, nil
	}

	resp, err := movieClient.GetAllGenres(ctx, &moviev1.GetAllGenresRequest{})
	if err != nil {
		return 0, false, err
	}

	for _, genre := range resp.GetGenres() {
		if genre == nil {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(genre.GetTitle()), value) {
			return genre.GetId(), true, nil
		}
	}

	return 0, false, nil
}

func orderMovieCardsByIDOrder(movieIDs []int64, movies []*moviev1.MovieCard) []*moviev1.MovieCard {
	byID := make(map[int64]*moviev1.MovieCard, len(movies))
	for _, m := range movies {
		if m == nil {
			continue
		}

		byID[m.GetId()] = m
	}

	out := make([]*moviev1.MovieCard, 0, len(movieIDs))
	for _, id := range movieIDs {
		if m, ok := byID[id]; ok {
			out = append(out, m)
		}
	}

	return out
}
