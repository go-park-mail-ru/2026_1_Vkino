package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

// GetEpisodesByMovieID Получает все эпизоды по movie_id
func (m *MovieUsecase) GetEpisodesByMovieID(ctx context.Context, movieID int64) (domain.MovieEpisodesResponse, error) {
	if movieID <= 0 {
		return domain.MovieEpisodesResponse{}, domain.ErrInvalidMovieID
	}

	episodes, err := m.movieRepo.GetEpisodesByMovieID(ctx, movieID)
	if err != nil {
		return domain.MovieEpisodesResponse{}, err
	}

	return episodes, nil
}

// GetEpisodePlayback Заполняет ссылку на видео в S3 и выставляет позицию просмотра
func (m *MovieUsecase) GetEpisodePlayback(ctx context.Context, episodeID, userID int64) (domain.EpisodePlaybackResponse, error) {
	if episodeID <= 0 {
		return domain.EpisodePlaybackResponse{}, domain.ErrInvalidEpisodeID
	}

	playback, err := m.movieRepo.GetEpisodePlayback(ctx, episodeID)
	if err != nil {
		return domain.EpisodePlaybackResponse{}, err
	}

	playbackURL, err := m.videoStorage.PresignGetObject(ctx, playback.PlaybackURL, 0)
	if err != nil {
		return domain.EpisodePlaybackResponse{}, err
	}
	playback.PlaybackURL = playbackURL

	if userID > 0 {
		positionSeconds, err := m.movieRepo.GetWatchProgress(ctx, userID, episodeID)
		if err != nil {
			return domain.EpisodePlaybackResponse{}, err
		}

		playback.PositionSeconds = positionSeconds
	}

	return playback, nil
}

// GetEpisodeProgress Получает позицию просмотра на которой остановились
func (m *MovieUsecase) GetEpisodeProgress(ctx context.Context, userID, episodeID int64) (domain.WatchProgressResponse, error) {
	if episodeID <= 0 {
		return domain.WatchProgressResponse{}, domain.ErrInvalidEpisodeID
	}

	positionSeconds, err := m.movieRepo.GetWatchProgress(ctx, userID, episodeID)
	if err != nil {
		return domain.WatchProgressResponse{}, err
	}

	return domain.WatchProgressResponse{
		EpisodeID:       episodeID,
		PositionSeconds: positionSeconds,
	}, nil
}

// SaveEpisodeProgress Записываем позицию просмотра
func (m *MovieUsecase) SaveEpisodeProgress(ctx context.Context, userID, episodeID int64, positionSeconds int) error {
	if episodeID <= 0 {
		return domain.ErrInvalidEpisodeID
	}

	if positionSeconds < 0 {
		return domain.ErrInvalidWatchProgress
	}

	return m.movieRepo.UpsertWatchProgress(ctx, userID, episodeID, positionSeconds)
}
