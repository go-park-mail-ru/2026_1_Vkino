package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetActorByID(ctx context.Context, actorID int64) (domain2.ActorResponse, error) {
	if actorID <= 0 {
		return domain2.ActorResponse{}, domain2.ErrInvalidActorID
	}

	actor, err := u.movieRepo.GetActorByID(ctx, actorID)
	if err != nil {
		return domain2.ActorResponse{}, err
	}

	return u.buildActorResponse(ctx, actor)
}
