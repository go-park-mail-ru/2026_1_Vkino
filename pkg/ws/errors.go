package ws

import "errors"

var (
	ErrNilUpgrader      = errors.New("ws upgrader is nil")
	ErrNilHub           = errors.New("ws hub is nil")
	ErrNilClient        = errors.New("ws client is nil")
	ErrNilConn          = errors.New("ws conn is nil")
	ErrClientIDRequired = errors.New("ws client id is required")
	ErrClientExists     = errors.New("ws client already subscribed")
	ErrHubClosed        = errors.New("ws hub is closed")
	ErrClientClosed     = errors.New("ws client is closed")
	ErrSendBufferFull   = errors.New("ws client send buffer is full")
)
