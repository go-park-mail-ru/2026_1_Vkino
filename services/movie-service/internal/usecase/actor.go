package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
)

func (u *MovieUsecase) GetActorByID(ctx context.Context, actorID int64) (domain.ActorResponse, error) {
	if actorID <= 0 {
		return domain.ActorResponse{}, domain.ErrInvalidActorID
	}

	actor, err := u.movieRepo.GetActorByID(ctx, actorID)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	return u.buildActorResponse(ctx, actor)
}
