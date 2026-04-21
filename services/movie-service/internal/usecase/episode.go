package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
)

func (u *MovieUsecase) GetEpisodePlayback(ctx context.Context, episodeID int64) (domain.EpisodePlaybackResponse, error) {
	if episodeID <= 0 {
		return domain.EpisodePlaybackResponse{}, domain.ErrInvalidEpisodeID
	}

	episode, err := u.movieRepo.GetEpisodePlayback(ctx, episodeID)
	if err != nil {
		return domain.EpisodePlaybackResponse{}, err
	}

	playbackURL, err := u.presignVideo(ctx, episode.VideoFileKey)
	if err != nil {
		return domain.EpisodePlaybackResponse{}, err
	}

	return domain.EpisodePlaybackResponse{
		EpisodeID:   episode.ID,
		PlaybackURL: playbackURL,
		DurationSec: episode.DurationSec,
	}, nil
}

func (u *MovieUsecase) GetEpisodeProgress(
	ctx context.Context,
	userID, episodeID int64,
) (domain.EpisodeProgressResponse, error) {
	if episodeID <= 0 {
		return domain.EpisodeProgressResponse{}, domain.ErrInvalidEpisodeID
	}
	if userID <= 0 {
		return domain.EpisodeProgressResponse{}, domain.ErrInternal
	}

	progress, err := u.movieRepo.GetEpisodeProgress(ctx, userID, episodeID)
	if err != nil {
		return domain.EpisodeProgressResponse{}, err
	}

	return domain.EpisodeProgressResponse{
		EpisodeID:   progress.EpisodeID,
		PositionSec: progress.PositionSec,
	}, nil
}

func (u *MovieUsecase) SaveEpisodeProgress(
	ctx context.Context,
	userID, episodeID, positionSec int64,
) (domain.EpisodeProgressResponse, error) {
	if episodeID <= 0 {
		return domain.EpisodeProgressResponse{}, domain.ErrInvalidEpisodeID
	}
	if userID <= 0 {
		return domain.EpisodeProgressResponse{}, domain.ErrInternal
	}
	if positionSec < 0 {
		return domain.EpisodeProgressResponse{}, domain.ErrInvalidWatchProgress
	}

	progress, err := u.movieRepo.SaveEpisodeProgress(ctx, userID, episodeID, positionSec)
	if err != nil {
		return domain.EpisodeProgressResponse{}, fmt.Errorf("%w: save episode progress: %v", domain.ErrInternal, err)
	}

	return domain.EpisodeProgressResponse{
		EpisodeID:   progress.EpisodeID,
		PositionSec: progress.PositionSec,
	}, nil
}
