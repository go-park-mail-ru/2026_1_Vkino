package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetEpisodePlayback(ctx context.Context,
	episodeID int64) (domain2.EpisodePlaybackResponse, error) {
	if episodeID <= 0 {
		return domain2.EpisodePlaybackResponse{}, domain2.ErrInvalidEpisodeID
	}

	episode, err := u.movieRepo.GetEpisodePlayback(ctx, episodeID)
	if err != nil {
		return domain2.EpisodePlaybackResponse{}, err
	}

	playbackURL, err := u.presignVideo(ctx, episode.VideoFileKey)
	if err != nil {
		return domain2.EpisodePlaybackResponse{}, err
	}

	return domain2.EpisodePlaybackResponse{
		EpisodeID:       episode.ID,
		MovieID:         episode.MovieID,
		SeasonNumber:    episode.SeasonNumber,
		EpisodeNumber:   episode.EpisodeNumber,
		Title:           episode.Title,
		DurationSeconds: episode.DurationSeconds,
		PlaybackURL:     playbackURL,
	}, nil
}

func (u *MovieUsecase) GetEpisodeProgress(
	ctx context.Context,
	userID, episodeID int64,
) (domain2.EpisodeProgressResponse, error) {
	if episodeID <= 0 {
		return domain2.EpisodeProgressResponse{}, domain2.ErrInvalidEpisodeID
	}

	if userID <= 0 {
		return domain2.EpisodeProgressResponse{}, domain2.ErrInternal
	}

	progress, err := u.movieRepo.GetEpisodeProgress(ctx, userID, episodeID)
	if err != nil {
		return domain2.EpisodeProgressResponse{}, err
	}

	return domain2.EpisodeProgressResponse(progress), nil
}

func (u *MovieUsecase) SaveEpisodeProgress(
	ctx context.Context,
	userID, episodeID, positionSec int64,
) (domain2.EpisodeProgressResponse, error) {
	if episodeID <= 0 {
		return domain2.EpisodeProgressResponse{}, domain2.ErrInvalidEpisodeID
	}

	if userID <= 0 {
		return domain2.EpisodeProgressResponse{}, domain2.ErrInternal
	}

	if positionSec < 0 {
		return domain2.EpisodeProgressResponse{}, domain2.ErrInvalidWatchProgress
	}

	progress, err := u.movieRepo.SaveEpisodeProgress(ctx, userID, episodeID, positionSec)
	if err != nil {
		return domain2.EpisodeProgressResponse{}, fmt.Errorf("%w: save episode progress: %v", domain2.ErrInternal, err)
	}

	return domain2.EpisodeProgressResponse(progress), nil
}
