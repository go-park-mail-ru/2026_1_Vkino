package usecase

import (
	"context"
	"fmt"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *supportUsecase) GetTicketStatistics(
	ctx context.Context,
	actorUserID int64,
	_ domain.GetSupportTicketStatisticsRequest,
) (domain.SupportTicketStatisticsResponse, error) {
	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain.SupportTicketStatisticsResponse{}, domain.ErrInvalidToken
	}

	if !isStaff(role) {
		return domain.SupportTicketStatisticsResponse{}, domain.ErrAccessDenied
	}

	stats, err := u.supportRepo.GetTicketStatistics(ctx, allowedCategoriesForRole(role))
	if err != nil {
		return domain.SupportTicketStatisticsResponse{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return *stats, nil
}
