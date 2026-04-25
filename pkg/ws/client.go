package ws

import (
	"context"
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		case payload := <-c.send:
			if err := c.conn.Write(ctx, payload); err != nil {
				return err
			}
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
