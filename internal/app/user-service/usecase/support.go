package usecase

import (
	"slices"
	"sync"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type supportUsecase struct {
	supportRepo      repository.SupportRepo
	userRepo         repository.UserRepo
	supportFileStore storage.FileStorage
	clockService     clocksvc.Service
	broker           *ticketBroker
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

func isAdmin(role string) bool {
	return role == "admin"
}

func isValidTicketCategory(category string) bool {
	if category == "" {
		return true
	}

	_, ok := ticketCategoryToSupportLine[category]

	return ok
}

func isValidTicketStatus(status string) bool {
	if status == "" {
		return true
	}

	switch status {
	case "open", "in_progress", "waiting_user", "resolved", "closed":
		return true
	default:
		return false
	}
}

func isValidSupportLine(line int64) bool {
	if line == 0 {
		return true
	}

	return line == 1 || line == 2
}

func supportLineForCategory(category string) int64 {
	return ticketCategoryToSupportLine[category]
}

func allowedCategoriesForRole(role string) []string {
	switch role {
	case "support_l1":
		return []string{"bug", "complaint", "question"}
	case "support_l2":
		return []string{"feature", "other"}
	default:
		return nil
	}
}

func canAccessCategory(role, category string) bool {
	allowedCategories := allowedCategoriesForRole(role)
	if len(allowedCategories) == 0 {
		return true
	}

	return slices.Contains(allowedCategories, category)
}

func isTerminalTicketStatus(status string) bool {
	return status == "resolved" || status == "closed"
}

var ticketCategoryToSupportLine = map[string]int64{
	"bug":       1,
	"complaint": 1,
	"question":  1,
	"feature":   2,
	"other":     2,
}
