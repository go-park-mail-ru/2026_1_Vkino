package ws

import (
	"context"
	"net/http"
)

// Conn abstracts a websocket connection implementation.
// A concrete adapter may wrap gorilla/websocket, nhooyr/websocket or another library.
type Conn interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, payload []byte) error
	Close() error
}

// Upgrader abstracts the HTTP upgrade handshake.
type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error)
}
