package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

type RoomEventBroker struct {
	mu          sync.RWMutex
	nextID      int64
	subscribers map[int64]map[int64]chan domain.RoomEvent
}

func NewRoomEventBroker() *RoomEventBroker {
	return &RoomEventBroker{
		subscribers: make(map[int64]map[int64]chan domain.RoomEvent),
	}
}

func (b *RoomEventBroker) Publish(_ context.Context, event domain.RoomEvent) error {
	b.mu.RLock()
	roomSubs := b.subscribers[event.RoomID]

	targets := make([]chan domain.RoomEvent, 0, len(roomSubs))
	for _, ch := range roomSubs {
		targets = append(targets, ch)
	}
	b.mu.RUnlock()

	for _, ch := range targets {
		select {
		case ch <- event:
		default:
		}
	}

	return nil
}

func (b *RoomEventBroker) Subscribe(_ context.Context, roomID int64) (<-chan domain.RoomEvent, func(), error) {
	if roomID <= 0 {
		return nil, nil, fmt.Errorf("invalid room id")
	}

	ch := make(chan domain.RoomEvent, 16)

	b.mu.Lock()
	b.nextID++
	subID := b.nextID

	if _, ok := b.subscribers[roomID]; !ok {
		b.subscribers[roomID] = make(map[int64]chan domain.RoomEvent)
	}
	b.subscribers[roomID][subID] = ch
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		roomSubs, ok := b.subscribers[roomID]
		if !ok {
			return
		}

		if subCh, ok := roomSubs[subID]; ok {
			delete(roomSubs, subID)
			close(subCh)
		}

		if len(roomSubs) == 0 {
			delete(b.subscribers, roomID)
		}
	}

	return ch, unsubscribe, nil
}
