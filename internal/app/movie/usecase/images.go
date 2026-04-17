package usecase

import "context"

func (m *MovieUsecase) presignCardURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	return m.cardStorage.PresignGetObject(ctx, key, 0)
}

func (m *MovieUsecase) presignPosterURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	return m.posterStorage.PresignGetObject(ctx, key, 0)
}

func (m *MovieUsecase) presignActorURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	return m.actorStorage.PresignGetObject(ctx, key, 0)
}

func (m *MovieUsecase) presignVideoURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	return m.videoStorage.PresignGetObject(ctx, key, 0)
}
