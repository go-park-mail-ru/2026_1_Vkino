package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/usecase"
)

type userRepoStub struct {
	resp  domain.ProfileResponse
	err   error
	gotID int64
}

func (s *userRepoStub) GetProfileByID(_ context.Context, id int64) (domain.ProfileResponse, error) {
	s.gotID = id

	if s.err != nil {
		return domain.ProfileResponse{}, s.err
	}

	return s.resp, nil
}

func TestProfileUsecase_GetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userID   int64
		repoResp domain.ProfileResponse
		repoErr  error
		wantResp domain.ProfileResponse
		wantErr  error
	}{
		{
			name:   "success",
			userID: 42,
			repoResp: domain.ProfileResponse{
				Email: "user@example.com",
			},
			wantResp: domain.ProfileResponse{
				Email: "user@example.com",
			},
		},
		{
			name:    "user not found",
			userID:  42,
			repoErr: domain.ErrUserNotFound,
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &userRepoStub{
				resp: tt.repoResp,
				err:  tt.repoErr,
			}

			u := usecase.NewProfileUsecase(repo)

			got, err := u.GetProfile(context.Background(), tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if got != tt.wantResp {
				t.Fatalf("expected response %+v, got %+v", tt.wantResp, got)
			}

			if repo.gotID != tt.userID {
				t.Fatalf("expected user id %d, got %d", tt.userID, repo.gotID)
			}
		})
	}
}
