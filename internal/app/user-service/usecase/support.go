package usecase

import (
	"sync"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
)

type supportUsecase struct {
	supportRepo repository.SupportRepo
	userRepo    repository.UserRepo
	broker      *ticketBroker
}

type ticketBroker struct {
	mu          sync.RWMutex
	subscribers map[int64][]chan domain.SupportTicketEventResponse
}

func newTicketBroker() *ticketBroker {
	return &ticketBroker{
		subscribers: make(map[int64][]chan domain.SupportTicketEventResponse),
	}
}

func (b *ticketBroker) subscribe(ticketID int64) (<-chan domain.SupportTicketEventResponse, func()) {
	ch := make(chan domain.SupportTicketEventResponse, 16)

	b.mu.Lock()
	b.subscribers[ticketID] = append(b.subscribers[ticketID], ch)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs := b.subscribers[ticketID]
		for i, sub := range subs {
			if sub == ch {
				b.subscribers[ticketID] = append(subs[:i], subs[i+1:]...)

				break
			}
		}

		close(ch)
	}

	return ch, unsubscribe
}

func (b *ticketBroker) publish(ticketID int64, event domain.SupportTicketEventResponse) {
	b.mu.RLock()
	subs := make([]chan domain.SupportTicketEventResponse, len(b.subscribers[ticketID]))
	copy(subs, b.subscribers[ticketID])
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func isStaff(role string) bool {
	return role == "support_l1" || role == "support_l2" || role == "admin"
}
