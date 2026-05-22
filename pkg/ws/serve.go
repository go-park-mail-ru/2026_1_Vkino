package ws

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type ClientIDFunc func(r *http.Request) (int64, error)

type ConnectHandler func(ctx context.Context, client *Client) error

type MessageHandler func(ctx context.Context, client *Client, payload []byte) error

type CloseHandler func(ctx context.Context, client *Client, err error)

type ServeOptions struct {
	SendBuffer int
	ClientID   ClientIDFunc
	OnConnect  ConnectHandler
	OnMessage  MessageHandler
	OnClose    CloseHandler
}

func ServeWS(upgrader Upgrader, hub *Hub, opts ServeOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client, ok := setupWSClient(w, r, upgrader, hub, opts)
		if !ok {
			return
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		defer func() {
			hub.Unsubscribe(client.ID())
		}()

		if opts.OnConnect != nil {
			if err := opts.OnConnect(ctx, client); err != nil {
				notifyClose(opts.OnClose, ctx, client, err)

				return
			}
		}

		writeErrCh := startWriteLoop(ctx, client)

		readErr := readLoop(ctx, client, opts.OnMessage)

		cancel()

		writeErr := <-writeErrCh
		closeErr := firstNonNil(readErr, suppressContextError(writeErr))
		notifyClose(opts.OnClose, ctx, client, closeErr)
	}
}

func setupWSClient(
	w http.ResponseWriter,
	r *http.Request,
	upgrader Upgrader,
	hub *Hub,
	opts ServeOptions,
) (*Client, bool) {
	if !validateServeWSDeps(w, upgrader, hub) {
		return nil, false
	}

	conn, err := upgrader.Upgrade(w, r)
	if err != nil {
		return nil, false
	}

	client, err := buildWSClient(r, conn, opts)
	if err != nil {
		_ = conn.Close()

		http.Error(w, err.Error(), http.StatusBadRequest)

		return nil, false
	}

	if err := hub.Subscribe(client); err != nil {
		_ = client.Close()

		http.Error(w, err.Error(), subscribeHTTPStatus(err))

		return nil, false
	}

	return client, true
}

func validateServeWSDeps(w http.ResponseWriter, upgrader Upgrader, hub *Hub) bool {
	if upgrader == nil {
		http.Error(w, ErrNilUpgrader.Error(), http.StatusInternalServerError)

		return false
	}

	if hub == nil {
		http.Error(w, ErrNilHub.Error(), http.StatusInternalServerError)

		return false
	}

	return true
}

func buildWSClient(r *http.Request, conn Conn, opts ServeOptions) (*Client, error) {
	clientID, err := resolveClientID(r, opts.ClientID)
	if err != nil {
		return nil, err
	}

	return NewClient(clientID, conn, opts.SendBuffer)
}

func subscribeHTTPStatus(err error) int {
	if errors.Is(err, ErrHubClosed) {
		return http.StatusServiceUnavailable
	}

	return http.StatusConflict
}

func startWriteLoop(ctx context.Context, client *Client) chan error {
	writeErrCh := make(chan error, 1)

	go func() {
		writeErrCh <- client.WriteLoop(ctx)
	}()

	return writeErrCh
}

func readLoop(ctx context.Context, client *Client, onMessage MessageHandler) error {
	for {
		payload, err := client.Conn().Read(ctx)
		if err != nil {
			return err
		}

		if onMessage == nil {
			continue
		}

		if err := onMessage(ctx, client, payload); err != nil {
			return err
		}
	}
}

func notifyClose(handler CloseHandler, ctx context.Context, client *Client, err error) {
	if handler != nil {
		handler(ctx, client, suppressContextError(err))
	}
}

func resolveClientID(r *http.Request, fn ClientIDFunc) (int64, error) {
	if fn == nil {
		return 0, ErrClientIDRequired
	}

	return fn(r)
}

func suppressContextError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}

	return err
}

func firstNonNil(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

type Group struct {
	hubs map[string]*Hub
	mu   sync.RWMutex
}

func NewGroup() *Group {
	return &Group{
		hubs: make(map[string]*Hub),
	}
}

func (g *Group) Hub(key string) *Hub {
	g.mu.RLock()
	hub, ok := g.hubs[key]
	g.mu.RUnlock()

	if ok {
		return hub
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if hub, ok = g.hubs[key]; ok {
		return hub
	}

	hub = NewHub()
	g.hubs[key] = hub

	return hub
}

func (g *Group) Close() error {
	g.mu.Lock()

	hubs := make([]*Hub, 0, len(g.hubs))
	for key, hub := range g.hubs {
		delete(g.hubs, key)
		hubs = append(hubs, hub)
	}
	g.mu.Unlock()

	for _, hub := range hubs {
		_ = hub.Close()
	}

	return nil
}
