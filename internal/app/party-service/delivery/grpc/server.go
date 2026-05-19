package grpc

import (
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/usecase"
)

type Server struct {
	partyv1.UnimplementedPartyServiceServer

	usecase     usecase.Usecase
	authClient  authv1.AuthServiceClient
	movieClient moviev1.MovieServiceClient
	userClient  userv1.UserServiceClient
}

func NewServer(
	u usecase.Usecase,
	authClient authv1.AuthServiceClient,
	movieClient moviev1.MovieServiceClient,
	userClient userv1.UserServiceClient,
) *Server {
	return &Server{
		usecase:     u,
		authClient:  authClient,
		movieClient: movieClient,
		userClient:  userClient,
	}
}
