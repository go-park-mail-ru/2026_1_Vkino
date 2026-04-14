package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

func (m *MovieUsecase) GetActorByID(ctx context.Context, id int64) (domain.ActorResponse, error) {
	actor, err := m.movieRepo.GetActorByID(ctx, id)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	actor.PictureFileKey, err = m.presignActorURL(ctx, actor.PictureFileKey)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	for i := range actor.Movies {
		actor.Movies[i].ImgUrl, err = m.presignCardURL(ctx, actor.Movies[i].ImgUrl)
		if err != nil {
			return domain.ActorResponse{}, err
		}
	}

	return actor, nil
}
