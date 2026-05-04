package routes

import (
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
