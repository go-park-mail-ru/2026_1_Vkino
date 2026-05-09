package usecase

import (
	"context"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

const (
	minMovieRating     = 0
	maxMovieRating     = 10
	maxMovieCommentLen = 4096
)

func (u *UserUsecase) SetMovieRating(
	ctx context.Context,
	userID, movieID int64,
	rating float64,
) (domain.MovieRatingResponse, error) {
	ratedMovie, err := u.SetMovieReview(ctx, userID, movieID, &rating, nil)
	if err != nil {
		return domain.MovieRatingResponse{}, err
	}

	if ratedMovie.Rating == nil {
		return domain.MovieRatingResponse{}, domain.ErrInternal
	}

	return domain.MovieRatingResponse{
		MovieID: ratedMovie.MovieID,
		Rating:  *ratedMovie.Rating,
	}, nil
}

func (u *UserUsecase) SetMovieReview(
	ctx context.Context,
	userID, movieID int64,
	rating *float64,
	comment *string,
) (domain.MovieReviewResponse, error) {
	if movieID <= 0 {
		return domain.MovieReviewResponse{}, domain.ErrInvalidMovieID
	}

	if rating != nil && (*rating < minMovieRating || *rating > maxMovieRating) {
		return domain.MovieReviewResponse{}, domain.ErrInvalidMovieRating
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.MovieReviewResponse{}, domain.ErrInvalidToken
	}

	normalizedComment, err := normalizeMovieReviewComment(comment)
	if err != nil {
		return domain.MovieReviewResponse{}, err
	}

	if rating == nil && normalizedComment == nil {
		return domain.MovieReviewResponse{}, domain.ErrInvalidMovieReview
	}

	return u.userRepo.SetMovieReview(ctx, userID, movieID, rating, normalizedComment)
}

func (u *UserUsecase) DeleteMovieReview(ctx context.Context, userID, movieID int64) error {
	if movieID <= 0 {
		return domain.ErrInvalidMovieID
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.ErrInvalidToken
	}

	return u.userRepo.DeleteMovieReview(ctx, userID, movieID)
}

func (u *UserUsecase) SetReviewReaction(
	ctx context.Context,
	userID, reviewID int64,
	reaction string,
) (domain.ReviewReactionResponse, error) {
	if reviewID <= 0 {
		return domain.ReviewReactionResponse{}, domain.ErrInvalidReviewID
	}

	normalizedReaction := strings.ToLower(strings.TrimSpace(reaction))
	if normalizedReaction != "like" && normalizedReaction != "dislike" {
		return domain.ReviewReactionResponse{}, domain.ErrInvalidReviewVote
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.ReviewReactionResponse{}, domain.ErrInvalidToken
	}

	if err := u.userRepo.SetReviewReaction(ctx, userID, reviewID, normalizedReaction); err != nil {
		return domain.ReviewReactionResponse{}, err
	}

	return domain.ReviewReactionResponse{
		ReviewID: reviewID,
		Reaction: normalizedReaction,
	}, nil
}

func (u *UserUsecase) DeleteReviewReaction(ctx context.Context, userID, reviewID int64) error {
	if reviewID <= 0 {
		return domain.ErrInvalidReviewID
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.ErrInvalidToken
	}

	return u.userRepo.DeleteReviewReaction(ctx, userID, reviewID)
}

func normalizeMovieReviewComment(comment *string) (*string, error) {
	if comment == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*comment)
	if trimmed == "" {
		return nil, nil
	}

	if len(trimmed) > maxMovieCommentLen {
		return nil, domain.ErrInvalidMovieComment
	}

	return &trimmed, nil
}
