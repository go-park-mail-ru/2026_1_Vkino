package domain

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidSearchQuery   = errors.New("invalid search query")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidMovieID       = errors.New("invalid movie id")
	ErrInvalidMovieRating   = errors.New("invalid movie rating")
	ErrInvalidMovieReview   = errors.New("invalid movie review payload")
	ErrInvalidMovieComment  = errors.New("invalid movie comment")
	ErrInvalidReviewID      = errors.New("invalid review id")
	ErrInvalidReviewVote    = errors.New("invalid review vote")
	ErrInvalidBirthdate     = errors.New("invalid birthdate")
	ErrInvalidAvatar        = errors.New("invalid avatar")
	ErrAlreadyFriends       = errors.New("already friends")
	ErrFriendNotFound       = errors.New("friend not found")
	ErrSelfFriendship       = errors.New("self friendship is forbidden")
	ErrMovieReviewNotFound  = errors.New("movie review not found")
	ErrSelfReviewVote       = errors.New("self review vote is forbidden")
	ErrInvalidRequestStatus = errors.New("invalid friend request status")
	ErrInternal             = errors.New("internal error")

	ErrTicketNotFound            = errors.New("ticket not found")
	ErrAccessDenied              = errors.New("access denied")
	ErrInvalidTicketID           = errors.New("invalid ticket id")
	ErrInvalidTicketPayload      = errors.New("invalid ticket payload")
	ErrInvalidMessage            = errors.New("invalid message payload")
	ErrInvalidSupportFilePayload = errors.New("invalid support file payload")
)
