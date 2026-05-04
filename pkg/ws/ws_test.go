package ws

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var errClosedByPeer = errors.New("closed by peer")

func TestHubSubscribeBroadcastAndUnsubscribe(t *testing.T) {
	hub := NewHub()

	clientA, err := NewClient(1, newStubConn(), 2)
	require.NoError(t, err)

	clientB, err := NewClient(2, newStubConn(), 2)
	require.NoError(t, err)

	require.NoError(t, hub.Subscribe(clientA))
	require.NoError(t, hub.Subscribe(clientB))
	require.Equal(t, 2, hub.Len())

	delivered := hub.Broadcast([]byte("hello"))
	require.Equal(t, 2, delivered)

	require.Equal(t, []byte("hello"), <-clientA.send)
	require.Equal(t, []byte("hello"), <-clientB.send)

	require.True(t, hub.Unsubscribe(1))
	require.Equal(t, 1, hub.Len())
}

func TestClientCloseStopsSend(t *testing.T) {
	client, err := NewClient(1, newStubConn(), 1)
	require.NoError(t, err)

	require.NoError(t, client.Close())
	require.ErrorIs(t, client.Send([]byte("x")), ErrClientClosed)
}

func TestServeWSLifecycle(t *testing.T) {
	hub := NewHub()
	conn := newStubConn()
	conn.reads = [][]byte{[]byte("msg-1")}
	conn.readErr = errClosedByPeer

	upgrader := &stubUpgrader{conn: conn}

	connected := false

	var disconnected error

	received := make(chan []byte, 1)

	handler := ServeWS(upgrader, hub, ServeOptions{
		SendBuffer: 1,
		ClientID: func(r *http.Request) (int64, error) {
			return 1, nil
		},
		OnConnect: func(ctx context.Context, client *Client) error {
			connected = true

			return client.Send([]byte("welcome"))
		},
		OnMessage: func(ctx context.Context, client *Client, payload []byte) error {
			received <- append([]byte(nil), payload...)

			return nil
		},
		OnClose: func(ctx context.Context, client *Client, err error) {
			disconnected = err
		},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, connected)
	require.Equal(t, []byte("msg-1"), <-received)
	require.Equal(t, [][]byte{[]byte("welcome")}, conn.writes)
	require.ErrorContains(t, disconnected, "closed by peer")
	require.Equal(t, 0, hub.Len())
	require.True(t, conn.closed)
}

func TestServeWSRequiresClientIDResolver(t *testing.T) {
	hub := NewHub()
	handler := ServeWS(&stubUpgrader{conn: newStubConn()}, hub, ServeOptions{})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), ErrClientIDRequired.Error())
}

type stubUpgrader struct {
	conn Conn
	err  error
}

func (u *stubUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error) {
	return u.conn, u.err
}

type stubConn struct {
	reads   [][]byte
	readErr error
	writes  [][]byte
	closed  bool
}

func newStubConn() *stubConn {
	return &stubConn{}
}

func (c *stubConn) Read(ctx context.Context) ([]byte, error) {
	if len(c.reads) > 0 {
		msg := c.reads[0]
		c.reads = c.reads[1:]

		return append([]byte(nil), msg...), nil
	}

	if c.readErr != nil {
		return nil, c.readErr
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return nil, context.Canceled
	}
}

func (c *stubConn) Write(ctx context.Context, payload []byte) error {
	c.writes = append(c.writes, append([]byte(nil), payload...))

	return nil
}

func (c *stubConn) Close() error {
	c.closed = true

	return nil
}
