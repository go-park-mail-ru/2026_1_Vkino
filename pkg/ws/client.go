package ws

import (
	"context"
	"errors"
	"sync"
)

const DefaultSendBuffer = 64

type Client struct {
	id   int64
	conn Conn
	send chan []byte

	done      chan struct{}
	closeOnce sync.Once
}

func NewClient(id int64, conn Conn, sendBuffer int) (*Client, error) {
	if conn == nil {
		return nil, ErrNilConn
	}

	if id == 0 {
		return nil, ErrClientIDRequired
	}

	if sendBuffer <= 0 {
		sendBuffer = DefaultSendBuffer
	}

	return &Client{
		id:   id,
		conn: conn,
		send: make(chan []byte, sendBuffer),
		done: make(chan struct{}),
	}, nil
}

func (c *Client) ID() int64 {
	return c.id
}

func (c *Client) Conn() Conn {
	return c.conn
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Send(payload []byte) error {
	select {
	case <-c.done:
		return ErrClientClosed
	default:
	}

	msg := append([]byte(nil), payload...)

	select {
	case <-c.done:
		return ErrClientClosed
	case c.send <- msg:
		return nil
	default:
		return ErrSendBufferFull
	}
}

func (c *Client) WriteLoop(ctx context.Context) error {
	for {
		if wrote, err := c.tryWritePending(ctx); wrote || err != nil {
			return err
		}

		err := c.waitForWriteOrShutdown(ctx)
		if err != nil {
			if errors.Is(err, ctx.Err()) {
				return c.flushOnContextDone(ctx)
			}

			return err
		}
	}
}

func (c *Client) Close() error {
	var err error

	c.closeOnce.Do(func() {
		close(c.done)
		err = c.conn.Close()
	})

	return err
}

func (c *Client) tryWritePending(ctx context.Context) (bool, error) {
	select {
	case payload := <-c.send:
		return true, c.conn.Write(ctx, payload)
	default:
		return false, nil
	}
}

func (c *Client) waitForWriteOrShutdown(ctx context.Context) error {
	select {
	case payload := <-c.send:
		return c.conn.Write(ctx, payload)
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) flushOnContextDone(ctx context.Context) error {
	select {
	case payload := <-c.send:
		return c.conn.Write(ctx, payload)
	default:
		return ctx.Err()
	}
}
