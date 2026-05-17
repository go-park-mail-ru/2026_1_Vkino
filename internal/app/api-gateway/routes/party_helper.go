package routes

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/api-gateway/domain"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

func collectOverviewMovieIDs(resp *partyv1.GetOverviewResponse) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)

	appendFromRooms := func(items []*partyv1.RoomCard) {
		for _, item := range items {
			if item == nil || item.GetPlayback() == nil {
				continue
			}

			movieID := item.GetPlayback().GetMovieId()
			if movieID == 0 {
				continue
			}

			if _, ok := seen[movieID]; ok {
				continue
			}

			seen[movieID] = struct{}{}
			result = append(result, movieID)
		}
	}

	appendFromRooms(resp.GetActiveRooms())
	appendFromRooms(resp.GetMyRooms())
	appendFromRooms(resp.GetFeaturedRooms())

	return result
}

func loadMovieImageURLs(
	ctx context.Context,
	movieClient moviev1.MovieServiceClient,
	movieIDs []int64,
) (map[int64]movieOverviewMeta, error) {
	result := make(map[int64]movieOverviewMeta, len(movieIDs))
	if len(movieIDs) == 0 {
		return result, nil
	}

	resp, err := movieClient.GetMoviesByIDs(ctx, &moviev1.GetMoviesByIDsRequest{MovieIds: movieIDs})
	if err != nil {
		return nil, err
	}

	for _, movie := range resp.GetMovies() {
		if movie == nil {
			continue
		}

		meta := movieOverviewMeta{}

		if value := strings.TrimSpace(movie.GetTitle()); value != "" {
			valueCopy := value
			meta.title = &valueCopy
		}

		if value := strings.TrimSpace(movie.GetImgUrl()); value != "" {
			valueCopy := value
			meta.imageURL = &valueCopy
		}

		result[movie.GetId()] = meta
	}

	return result, nil
}

type movieOverviewMeta struct {
	title    *string
	imageURL *string
}

func mapOverviewResponse(resp *partyv1.GetOverviewResponse,
	movieImages map[int64]movieOverviewMeta) domain.PartyOverviewResponse {
	return domain.PartyOverviewResponse{
		ActiveRooms:   mapRoomCardsHTTP(resp.GetActiveRooms(), movieImages),
		MyRooms:       mapRoomCardsHTTP(resp.GetMyRooms(), movieImages),
		FeaturedRooms: mapRoomCardsHTTP(resp.GetFeaturedRooms(), movieImages),
	}
}

func mapRoomCardsHTTP(items []*partyv1.RoomCard, movieImages map[int64]movieOverviewMeta) []domain.PartyRoomCardHTTP {
	result := make([]domain.PartyRoomCardHTTP, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}

		result = append(result, domain.PartyRoomCardHTTP{
			ID:                item.GetId(),
			Name:              item.GetName(),
			Visibility:        item.GetVisibility(),
			InviteLink:        item.GetInviteLink(),
			HostUserID:        item.GetHostUserId(),
			HostName:          item.GetHostName(),
			ParticipantsCount: item.GetParticipantsCount(),
			Playback:          mapPlaybackHTTP(item.GetPlayback(), movieImages),
			UpdatedAt:         item.GetUpdatedAt(),
		})
	}

	return result
}

func mapPlaybackHTTP(item *partyv1.PlaybackState,
	movieImages map[int64]movieOverviewMeta) *domain.PartyPlaybackStateHTTP {
	if item == nil {
		return nil
	}

	meta := movieImages[item.GetMovieId()]

	return &domain.PartyPlaybackStateHTTP{
		MovieID:         item.GetMovieId(),
		MovieTitle:      meta.title,
		EpisodeID:       item.GetEpisodeId(),
		ImgURL:          meta.imageURL,
		PlaybackURL:     item.GetPlaybackUrl(),
		DurationSeconds: item.GetDurationSeconds(),
		PositionSeconds: item.GetPositionSeconds(),
		Status:          item.GetStatus(),
		UpdatedAt:       item.GetUpdatedAt(),
	}
}

func parseRoomPathID(w http.ResponseWriter, r *http.Request, message string) (int64, bool) {
	value := strings.TrimSpace(r.PathValue("id"))
	if value == "" {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	value = strings.TrimPrefix(strings.ToLower(value), "id")

	roomID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		httppkg.ErrResponse(w, http.StatusBadRequest, message)

		return 0, false
	}

	return roomID, true
}
