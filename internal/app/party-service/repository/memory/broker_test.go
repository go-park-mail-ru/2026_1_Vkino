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

func TestRoomEventBrokerPublishDoesNotBlockOnSlowSubscriber(t *testing.T) {
	t.Parallel()

	const roomID int64 = 9

	const eventsCount = subscriberBufferSize + 1

	broker := NewRoomEventBroker()

	slowEvents, unsubscribeSlow, err := broker.Subscribe(context.Background(), roomID, 1)
	if err != nil {
		t.Fatalf("subscribe slow: %v", err)
	}
	defer unsubscribeSlow()

	fastEvents, unsubscribeFast, err := broker.Subscribe(context.Background(), roomID, 2)
	if err != nil {
		t.Fatalf("subscribe fast: %v", err)
	}
	defer unsubscribeFast()

	publishDone := make(chan error, 1)

	go func() {
		for {
			select {
			case <-publishDone:
				return
			case <-fastEvents:
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
	}()

	go func() {
		publishDone <- publishRoomEvents(broker, roomID, eventsCount)
	}()

	assertPublishCompletes(t, publishDone)
	assertBufferedEventCount(t, slowEvents, subscriberBufferSize)
}

func TestRoomEventBrokerPublishSkipsPlaybackEventForActor(t *testing.T) {
	t.Parallel()

	broker := NewRoomEventBroker()

	actorEvents, unsubscribeActor, err := broker.Subscribe(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("subscribe actor: %v", err)
	}
	defer unsubscribeActor()

	peerEvents, unsubscribePeer, err := broker.Subscribe(context.Background(), 42, 8)
	if err != nil {
		t.Fatalf("subscribe peer: %v", err)
	}
	defer unsubscribePeer()

	event := domain.RoomEvent{
		Type:        "sync_state",
		RoomID:      42,
		ActorUserID: 7,
		Playback: &domain.PlaybackState{
			MovieID: 10,
			Status:  "playing",
		},
	}

	if err = broker.Publish(context.Background(), event); err != nil {
		t.Fatalf("publish: %v", err)
	}

	assertNoEvent(t, actorEvents)
	assertEvent(t, peerEvents, event)
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

func assertBufferedEventCount(t *testing.T, events <-chan domain.RoomEvent, want int) {
	t.Helper()

	for i := range want {
		event := readEvent(t, events)
		if event.ActorUserID != int64(i) {
			t.Fatalf("unexpected buffered event order at index %d: got actor_user_id=%d want=%d", i, event.ActorUserID, i)
		}
	}

	select {
	case event := <-events:
		t.Fatalf("expected slow subscriber buffer to stop at %d events, got extra event %+v", want, event)
	default:
	}
}

func assertPublishCompletes(t *testing.T, done <-chan error) {
	t.Helper()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("publish returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("publish blocked on slow subscriber")
	}
}

func assertEvent(t *testing.T, ch <-chan domain.RoomEvent, want domain.RoomEvent) {
	t.Helper()

	got := readEvent(t, ch)

	if got.Type != want.Type || got.RoomID != want.RoomID || got.ActorUserID != want.ActorUserID {
		t.Fatalf("unexpected event: got=%+v want=%+v", got, want)
	}
}

func assertNoEvent(t *testing.T, ch <-chan domain.RoomEvent) {
	t.Helper()

	select {
	case event := <-ch:
		t.Fatalf("unexpected event: %+v", event)
	case <-time.After(100 * time.Millisecond):
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
