package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *supportUsecase) GetTickets(
	ctx context.Context,
	actorUserID int64,
	req domain2.GetSupportTicketsRequest,
) ([]domain2.SupportTicketResponse, error) {
	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return nil, domain2.ErrInvalidToken
	}

	userIDFilter := actorUserID
	if isStaff(role) {
		userIDFilter = 0
	} else {
		req.SupportLine = 0
	}

	tickets, err := u.supportRepo.GetTickets(ctx, userIDFilter, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return tickets, nil
}
