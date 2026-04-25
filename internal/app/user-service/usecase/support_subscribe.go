package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *supportUsecase) SubscribeTicket(
	ctx context.Context,
	actorUserID int64,
	req domain2.SubscribeSupportTicketRequest,
) (<-chan domain2.SupportTicketEventResponse, func(), error) {
	if req.TicketID <= 0 {
		return nil, nil, domain2.ErrInvalidTicketID
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return nil, nil, err
	}

	ch, unsubscribe := u.broker.subscribe(req.TicketID)

	return ch, unsubscribe, nil
}
