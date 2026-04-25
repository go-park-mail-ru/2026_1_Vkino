package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *supportUsecase) CreateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain2.CreateSupportTicketRequest,
) (domain2.SupportTicketResponse, error) {
	if req.Title == "" || req.Description == "" || req.Category == "" {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketID
	}

	ticket, err := u.supportRepo.CreateTicket(ctx, actorUserID, req)
	if err != nil {
		return domain2.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return *ticket, nil
}
