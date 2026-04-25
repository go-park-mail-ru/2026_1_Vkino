package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *supportUsecase) GetTicketStatistics(
	ctx context.Context,
	actorUserID int64,
	_ domain2.GetSupportTicketStatisticsRequest,
) (domain2.SupportTicketStatisticsResponse, error) {
	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain2.SupportTicketStatisticsResponse{}, domain2.ErrInvalidToken
	}

	if !isStaff(role) {
		return domain2.SupportTicketStatisticsResponse{}, domain2.ErrAccessDenied
	}

	stats, err := u.supportRepo.GetTicketStatistics(ctx)
	if err != nil {
		return domain2.SupportTicketStatisticsResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return *stats, nil
}
