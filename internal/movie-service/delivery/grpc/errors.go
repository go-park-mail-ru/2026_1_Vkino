package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	"google.golang.org/grpc/codes"
)

var movieGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrInvalidMovieID,
		domain.ErrInvalidActorID,
		domain.ErrInvalidEpisodeID,
		domain.ErrInvalidSelectionTitle,
		domain.ErrInvalidSearchQuery,
		domain.ErrInvalidWatchProgress,

		domain.ErrMovieNotFound,
		postgresrepo.ErrMovieNotFound,

		domain.ErrActorNotFound,
		postgresrepo.ErrActorNotFound,

		domain.ErrSelectionNotFound,
		postgresrepo.ErrSelectionNotFound,

		domain.ErrEpisodeNotFound,
		postgresrepo.ErrEpisodeNotFound,

		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrInvalidMovieID:        {Code: codes.InvalidArgument, Message: "invalid movie id"},
		domain.ErrInvalidActorID:        {Code: codes.InvalidArgument, Message: "invalid actor id"},
		domain.ErrInvalidEpisodeID:      {Code: codes.InvalidArgument, Message: "invalid episode id"},
		domain.ErrInvalidSelectionTitle: {Code: codes.InvalidArgument, Message: "invalid selection title"},
		domain.ErrInvalidSearchQuery:    {Code: codes.InvalidArgument, Message: "invalid search query"},
		domain.ErrInvalidWatchProgress:  {Code: codes.InvalidArgument, Message: "invalid watch progress"},

		domain.ErrMovieNotFound:       {Code: codes.NotFound, Message: "movie not found"},
		postgresrepo.ErrMovieNotFound: {Code: codes.NotFound, Message: "movie not found"},

		domain.ErrActorNotFound:       {Code: codes.NotFound, Message: "actor not found"},
		postgresrepo.ErrActorNotFound: {Code: codes.NotFound, Message: "actor not found"},

		domain.ErrSelectionNotFound:       {Code: codes.NotFound, Message: "selection not found"},
		postgresrepo.ErrSelectionNotFound: {Code: codes.NotFound, Message: "selection not found"},

		domain.ErrEpisodeNotFound:       {Code: codes.NotFound, Message: "episode not found"},
		postgresrepo.ErrEpisodeNotFound: {Code: codes.NotFound, Message: "episode not found"},

		domain.ErrInternal: {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapError(err error) error {
	return movieGRPCErrorMapper.Map(err)
}
