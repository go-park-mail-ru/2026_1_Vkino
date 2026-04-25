package ws

import (
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]*Client
	closed  bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]*Client),
	}
}

func (h *Hub) Subscribe(client *Client) error {
	if client == nil {
		return ErrNilClient
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return ErrHubClosed
	}

	if _, exists := h.clients[client.ID()]; exists {
		return ErrClientExists
	}

	h.clients[client.ID()] = client

	return nil
}

func (h *Hub) Unsubscribe(clientID int64) bool {
	h.mu.Lock()

	client, ok := h.clients[clientID]
	if ok {
		delete(h.clients, clientID)
	}
	h.mu.Unlock()

	if ok {
		_ = client.Close()
	}

	return ok
}

func (h *Hub) Client(clientID int64) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[clientID]

	return client, ok
}

func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

func (h *Hub) Broadcast(payload []byte) int {
	h.mu.RLock()

	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}

	h.mu.RUnlock()

	delivered := 0

	for _, client := range clients {
		if err := client.Send(payload); err == nil {
			delivered++
		}
	}

	return delivered
}

func (h *Hub) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()

		return nil
	}

	h.closed = true

	clients := make([]*Client, 0, len(h.clients))
	for id, client := range h.clients {
		delete(h.clients, id)
		clients = append(clients, client)
	}
	h.mu.Unlock()

	for _, client := range clients {
		_ = client.Close()
	}

	return nil
}
