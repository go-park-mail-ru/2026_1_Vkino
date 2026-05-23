package memory

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

func TestRoomEventBrokerPublishDeliversToAllSubscribers(t *testing.T) {
	t.Parallel()

	broker := NewRoomEventBroker()

	first, unsubFirst, err := broker.Subscribe(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("subscribe first: %v", err)
	}
	defer unsubFirst()

	second, unsubSecond, err := broker.Subscribe(context.Background(), 42, 2)
	if err != nil {
		t.Fatalf("subscribe second: %v", err)
	}
	defer unsubSecond()

	event := domain.RoomEvent{
		Type:        "pause",
		RoomID:      42,
		ActorUserID: 7,
	}

	if err = broker.Publish(context.Background(), event); err != nil {
		t.Fatalf("publish: %v", err)
	}

	assertEvent(t, first, event)
	assertEvent(t, second, event)
}

func TestRoomEventBrokerPublishWaitsInsteadOfDropping(t *testing.T) {
	t.Parallel()

	const roomID int64 = 9

	const eventsCount = 17

	broker := NewRoomEventBroker()

	events, unsubscribe, err := broker.Subscribe(context.Background(), roomID, 1)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer unsubscribe()

	done := make(chan error, 1)

	go func() {
		done <- publishRoomEvents(broker, roomID, eventsCount)
	}()

	assertPublishBlocks(t, done)
	assertPublishedEventOrder(t, events, eventsCount)
	assertPublishCompletes(t, done)
}

func publishRoomEvents(broker *RoomEventBroker, roomID int64, eventsCount int) error {
	for i := range eventsCount {
		if err := broker.Publish(context.Background(), domain.RoomEvent{
			Type:        "sync_state",
			RoomID:      roomID,
			ActorUserID: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

func assertPublishBlocks(t *testing.T, done <-chan error) {
	t.Helper()

	select {
	case err := <-done:
		t.Fatalf("publish finished early, expected blocking after buffer fill: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
}

func assertPublishedEventOrder(t *testing.T, events <-chan domain.RoomEvent, eventsCount int) {
	t.Helper()

	for i := range eventsCount {
		event := readEvent(t, events)
		if event.ActorUserID != int64(i) {
			t.Fatalf("unexpected event order at index %d: got actor_user_id=%d want=%d", i, event.ActorUserID, i)
		}
	}
}

func assertPublishCompletes(t *testing.T, done <-chan error) {
	t.Helper()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("publish after drain: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("publish did not finish after subscriber drained events")
	}
}

func assertEvent(t *testing.T, ch <-chan domain.RoomEvent, want domain.RoomEvent) {
	t.Helper()

	got := readEvent(t, ch)

	if got.Type != want.Type || got.RoomID != want.RoomID || got.ActorUserID != want.ActorUserID {
		t.Fatalf("unexpected event: got=%+v want=%+v", got, want)
	}
}

func readEvent(t *testing.T, ch <-chan domain.RoomEvent) domain.RoomEvent {
	t.Helper()

	select {
	case event := <-ch:
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")

		return domain.RoomEvent{}
	}
}
