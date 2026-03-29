package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
	"github.com/google/uuid"
)

func (m *MovieUsecase) GetActorByID(ctx context.Context, id uuid.UUID) (domain.ActorResponse, error) {
	actor, err := m.movieRepo.GetActorByID(ctx, id)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	return actor, nil
}
