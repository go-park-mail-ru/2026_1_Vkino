package memory

import (
	"context"
	"errors"
	"math"
	"sync"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

const subscriberBufferSize = 16

var errInvalidRoomID = errors.New("invalid room id")

type RoomEventBroker struct {
	mu          sync.RWMutex
	nextID      int64
	subscribers map[int64]map[int64]roomSubscriber
}

type roomSubscriber struct {
	userID int64
	ch     chan domain.RoomEvent
}

type roomSubscriberSnapshot struct {
	userID int64
	ch     chan domain.RoomEvent
}

func NewRoomEventBroker() *RoomEventBroker {
	return &RoomEventBroker{
		subscribers: make(map[int64]map[int64]roomSubscriber),
	}
}

func (b *RoomEventBroker) Publish(ctx context.Context, event domain.RoomEvent) error {
	subscribers := b.roomSubscribers(event.RoomID)
	for _, subscriber := range subscribers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case subscriber.ch <- event:
		default:
		}
	}

	return nil
}

func (b *RoomEventBroker) Subscribe(
	_ context.Context,
	roomID int64,
	userID int64,
) (<-chan domain.RoomEvent, func(), error) {
	if roomID <= 0 || userID <= 0 {
		return nil, nil, errInvalidRoomID
	}

	ch := make(chan domain.RoomEvent, subscriberBufferSize)

	b.mu.Lock()
	b.nextID++
	subID := b.nextID

	if _, ok := b.subscribers[roomID]; !ok {
		b.subscribers[roomID] = make(map[int64]roomSubscriber)
	}

	b.subscribers[roomID][subID] = roomSubscriber{userID: userID, ch: ch}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		roomSubs, ok := b.subscribers[roomID]
		if !ok {
			return
		}

		if sub, ok := roomSubs[subID]; ok {
			delete(roomSubs, subID)
			close(sub.ch)
		}

		if len(roomSubs) == 0 {
			delete(b.subscribers, roomID)
		}
	}

	return ch, unsubscribe, nil
}

func (b *RoomEventBroker) ActiveUsers(roomID int64) int32 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	users := make(map[int64]struct{})
	for _, subscriber := range b.subscribers[roomID] {
		users[subscriber.userID] = struct{}{}
	}

	var count int32
	for range users {
		if count == math.MaxInt32 {
			return math.MaxInt32
		}

		count++
	}

	return count
}

func (b *RoomEventBroker) IsUserActive(roomID, userID int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, subscriber := range b.subscribers[roomID] {
		if subscriber.userID == userID {
			return true
		}
	}

	return false
}

func (b *RoomEventBroker) roomSubscribers(roomID int64) []roomSubscriberSnapshot {
	b.mu.RLock()
	defer b.mu.RUnlock()

	roomSubs := b.subscribers[roomID]
	if len(roomSubs) == 0 {
		return nil
	}

	snapshots := make([]roomSubscriberSnapshot, 0, len(roomSubs))
	for _, subscriber := range roomSubs {
		snapshots = append(snapshots, roomSubscriberSnapshot(subscriber))
	}

	return snapshots
}
