package routes

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	wspkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/ws"
)

const (
	websocketMagicGUID      = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	websocketOpcodeText     = 0x1
	websocketOpcodeClose    = 0x8
	websocketOpcodePing     = 0x9
	websocketOpcodePong     = 0xA
	maxWebSocketPayloadSize = 1 << 20
)

type supportWSUpgrader struct{}

func (supportWSUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (wspkg.Conn, error) {
	if !headerContainsToken(r.Header, "Connection", "Upgrade") || !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)

		return nil, fmt.Errorf("websocket upgrade required")
	}

	if strings.TrimSpace(r.Header.Get("Sec-WebSocket-Version")) != "13" {
		http.Error(w, "unsupported websocket version", http.StatusUpgradeRequired)

		return nil, fmt.Errorf("unsupported websocket version")
	}

	secKey := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if secKey == "" {
		http.Error(w, "missing websocket key", http.StatusBadRequest)

		return nil, fmt.Errorf("missing websocket key")
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket hijacking is not supported", http.StatusInternalServerError)

		return nil, fmt.Errorf("response writer does not support hijacking")
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijack websocket connection: %w", err)
	}

	accept := buildWebSocketAccept(secKey)

	if _, err = rw.WriteString(
		"HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + accept + "\r\n\r\n",
	); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("write websocket handshake: %w", err)
	}

	if err = rw.Flush(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("flush websocket handshake: %w", err)
	}

	return &supportWSConn{
		conn:   conn,
		reader: rw.Reader,
		writer: rw.Writer,
	}, nil
}

type supportWSConn struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *supportWSConn) Read(ctx context.Context) ([]byte, error) {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetReadDeadline)
	defer stopDeadline()

	for {
		opcode, payload, err := c.readFrame()
		if err != nil {
			return nil, err
		}

		switch opcode {
		case websocketOpcodeText:
			return payload, nil

		case websocketOpcodePing:
			if err = c.writeControlFrame(ctx, websocketOpcodePong, payload); err != nil {
				return nil, err
			}

		case websocketOpcodePong:
			continue

		case websocketOpcodeClose:
			_ = c.writeControlFrame(ctx, websocketOpcodeClose, nil)

			return nil, io.EOF

		default:
			return nil, fmt.Errorf("unsupported websocket opcode %d", opcode)
		}
	}
}

func (c *supportWSConn) Write(ctx context.Context, payload []byte) error {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetWriteDeadline)
	defer stopDeadline()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.writeFrameLocked(websocketOpcodeText, payload)
}

func (c *supportWSConn) Close() error {
	var err error

	c.closeOnce.Do(func() {
		err = c.conn.Close()
	})

	return err
}

func (c *supportWSConn) readFrame() (byte, []byte, error) {
	firstByte, err := c.reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	if firstByte&0x80 == 0 {
		return 0, nil, fmt.Errorf("fragmented websocket frames are not supported")
	}

	opcode := firstByte & 0x0F

	secondByte, err := c.reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	if secondByte&0x80 == 0 {
		return 0, nil, fmt.Errorf("client websocket frames must be masked")
	}

	payloadLength, err := readWebSocketPayloadLength(c.reader, secondByte&0x7F)
	if err != nil {
		return 0, nil, err
	}

	if payloadLength > maxWebSocketPayloadSize {
		return 0, nil, fmt.Errorf("websocket payload exceeds limit")
	}

	var maskKey [4]byte
	if _, err = io.ReadFull(c.reader, maskKey[:]); err != nil {
		return 0, nil, err
	}

	payload := make([]byte, payloadLength)
	if _, err = io.ReadFull(c.reader, payload); err != nil {
		return 0, nil, err
	}

	for idx := range payload {
		payload[idx] ^= maskKey[idx%len(maskKey)]
	}

	return opcode, payload, nil
}

func (c *supportWSConn) writeControlFrame(ctx context.Context, opcode byte, payload []byte) error {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetWriteDeadline)
	defer stopDeadline()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.writeFrameLocked(opcode, payload)
}

func (c *supportWSConn) writeFrameLocked(opcode byte, payload []byte) error {
	if err := c.writer.WriteByte(0x80 | opcode); err != nil {
		return err
	}

	payloadLength := len(payload)
	switch {
	case payloadLength <= 125:
		if err := c.writer.WriteByte(byte(payloadLength)); err != nil {
			return err
		}

	case payloadLength <= 65535:
		if err := c.writer.WriteByte(126); err != nil {
			return err
		}

		var rawLength [2]byte
		binary.BigEndian.PutUint16(rawLength[:], uint16(payloadLength))
		if _, err := c.writer.Write(rawLength[:]); err != nil {
			return err
		}

	default:
		if err := c.writer.WriteByte(127); err != nil {
			return err
		}

		var rawLength [8]byte
		binary.BigEndian.PutUint64(rawLength[:], uint64(payloadLength))
		if _, err := c.writer.Write(rawLength[:]); err != nil {
			return err
		}
	}

	if _, err := c.writer.Write(payload); err != nil {
		return err
	}

	return c.writer.Flush()
}

func headerContainsToken(header http.Header, key, want string) bool {
	for _, value := range header.Values(key) {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), want) {
				return true
			}
		}
	}

	return false
}

func buildWebSocketAccept(secKey string) string {
	hash := sha1.Sum([]byte(secKey + websocketMagicGUID))

	return base64.StdEncoding.EncodeToString(hash[:])
}

func readWebSocketPayloadLength(reader *bufio.Reader, payloadLengthCode byte) (int, error) {
	switch payloadLengthCode {
	case 126:
		var rawLength [2]byte
		if _, err := io.ReadFull(reader, rawLength[:]); err != nil {
			return 0, err
		}

		return int(binary.BigEndian.Uint16(rawLength[:])), nil

	case 127:
		var rawLength [8]byte
		if _, err := io.ReadFull(reader, rawLength[:]); err != nil {
			return 0, err
		}

		payloadLength := binary.BigEndian.Uint64(rawLength[:])
		if payloadLength > uint64(maxWebSocketPayloadSize) {
			return 0, fmt.Errorf("websocket payload exceeds limit")
		}

		return int(payloadLength), nil

	default:
		return int(payloadLengthCode), nil
	}
}

func watchConnDeadline(ctx context.Context, setDeadline func(time.Time) error) func() {
	if deadline, ok := ctx.Deadline(); ok {
		_ = setDeadline(deadline)
	} else {
		_ = setDeadline(time.Time{})
	}

	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			_ = setDeadline(time.Now())
		case <-done:
		}
	}()

	return func() {
		close(done)
		_ = setDeadline(time.Time{})
	}
}
