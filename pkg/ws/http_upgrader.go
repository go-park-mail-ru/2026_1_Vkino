package ws

import (
	"bufio"
	"context"
	"crypto/sha1" // #nosec G505 -- RFC 6455 requires SHA-1 for the WebSocket handshake accept key.
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	websocketMagicGUID      = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	websocketOpcodeText     = 0x1
	websocketOpcodeClose    = 0x8
	websocketOpcodePing     = 0x9
	websocketOpcodePong     = 0xA
	maxWebSocketPayloadSize = 1 << 20
	websocketFinalBit       = 0x80
	websocketOpcodeMask     = 0x0F
	websocketMaskBit        = 0x80
	websocketPayloadMask    = 0x7F
	websocketPayload126     = 126
	websocketPayload127     = 127
	websocketPayloadInline  = 125
)

var (
	errWebSocketUpgradeRequired      = errors.New("websocket upgrade required")
	errUnsupportedWebSocketVersion   = errors.New("unsupported websocket version")
	errMissingWebSocketKey           = errors.New("missing websocket key")
	errWebSocketHijackingUnsupported = errors.New("response writer does not support hijacking")
	errUnsupportedWebSocketOpcode    = errors.New("unsupported websocket opcode")
	errFragmentedWebSocketFrames     = errors.New("fragmented websocket frames are not supported")
	errUnmaskedWebSocketFrames       = errors.New("client websocket frames must be masked")
	errWebSocketPayloadTooLarge      = errors.New("websocket payload exceeds limit")
	errInlinePayloadLengthOutOfRange = errors.New("inline websocket payload length out of range")
)

type HTTPUpgrader struct{}

func (HTTPUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error) {
	secKey, err := validateUpgradeRequest(w, r)
	if err != nil {
		return nil, err
	}

	conn, rw, err := hijackUpgradeConnection(w)
	if err != nil {
		return nil, err
	}

	if err = writeHandshakeResponse(conn, rw, secKey); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return &httpConn{
		conn:   conn,
		reader: rw.Reader,
		writer: rw.Writer,
	}, nil
}

type httpConn struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *httpConn) Read(ctx context.Context) ([]byte, error) {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetReadDeadline)
	defer stopDeadline()

	for {
		opcode, payload, err := c.readFrame()
		if err != nil {
			return nil, err
		}

		message, done, err := c.handleFrame(ctx, opcode, payload)
		if done || err != nil {
			return message, err
		}
	}
}

func (c *httpConn) Write(ctx context.Context, payload []byte) error {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetWriteDeadline)
	defer stopDeadline()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.writeFrameLocked(websocketOpcodeText, payload)
}

func (c *httpConn) Close() error {
	var err error

	c.closeOnce.Do(func() {
		err = c.conn.Close()
	})

	return err
}

func (c *httpConn) readFrame() (byte, []byte, error) {
	opcode, payloadLength, err := c.readFrameHeader()
	if err != nil {
		return 0, nil, err
	}

	if payloadLength > maxWebSocketPayloadSize {
		return 0, nil, errWebSocketPayloadTooLarge
	}

	payload, err := c.readMaskedPayload(payloadLength)
	if err != nil {
		return 0, nil, err
	}

	return opcode, payload, nil
}

func (c *httpConn) writeControlFrame(ctx context.Context, opcode byte, payload []byte) error {
	stopDeadline := watchConnDeadline(ctx, c.conn.SetWriteDeadline)
	defer stopDeadline()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.writeFrameLocked(opcode, payload)
}

func (c *httpConn) writeFrameLocked(opcode byte, payload []byte) error {
	if err := c.writer.WriteByte(websocketFinalBit | opcode); err != nil {
		return err
	}

	if err := writePayloadLength(c.writer, len(payload)); err != nil {
		return err
	}

	if _, err := c.writer.Write(payload); err != nil {
		return err
	}

	return c.writer.Flush()
}

func validateUpgradeRequest(w http.ResponseWriter, r *http.Request) (string, error) {
	if !headerContainsToken(r.Header, "Connection", "Upgrade") ||
		!strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)

		return "", errWebSocketUpgradeRequired
	}

	if strings.TrimSpace(r.Header.Get("Sec-WebSocket-Version")) != "13" {
		http.Error(w, "unsupported websocket version", http.StatusUpgradeRequired)

		return "", errUnsupportedWebSocketVersion
	}

	secKey := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if secKey == "" {
		http.Error(w, "missing websocket key", http.StatusBadRequest)

		return "", errMissingWebSocketKey
	}

	return secKey, nil
}

func hijackUpgradeConnection(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket hijacking is not supported", http.StatusInternalServerError)

		return nil, nil, errWebSocketHijackingUnsupported
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, fmt.Errorf("hijack websocket connection: %w", err)
	}

	return conn, rw, nil
}

func writeHandshakeResponse(conn net.Conn, rw *bufio.ReadWriter, secKey string) error {
	accept := buildWebSocketAccept(secKey)

	if _, err := rw.WriteString(
		"HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + accept + "\r\n\r\n",
	); err != nil {
		_ = conn.Close()

		return fmt.Errorf("write websocket handshake: %w", err)
	}

	if err := rw.Flush(); err != nil {
		_ = conn.Close()

		return fmt.Errorf("flush websocket handshake: %w", err)
	}

	return nil
}

func (c *httpConn) readFrameHeader() (byte, int, error) {
	firstByte, err := c.reader.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if firstByte&websocketFinalBit == 0 {
		return 0, 0, errFragmentedWebSocketFrames
	}

	secondByte, err := c.reader.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if secondByte&websocketMaskBit == 0 {
		return 0, 0, errUnmaskedWebSocketFrames
	}

	payloadLength, err := readWebSocketPayloadLength(c.reader, secondByte&websocketPayloadMask)
	if err != nil {
		return 0, 0, err
	}

	return firstByte & websocketOpcodeMask, payloadLength, nil
}

func (c *httpConn) handleFrame(ctx context.Context, opcode byte, payload []byte) ([]byte, bool, error) {
	switch opcode {
	case websocketOpcodeText:
		return payload, true, nil
	case websocketOpcodePing:
		if err := c.writeControlFrame(ctx, websocketOpcodePong, payload); err != nil {
			return nil, true, err
		}

		return nil, false, nil
	case websocketOpcodePong:
		return nil, false, nil
	case websocketOpcodeClose:
		if err := c.writeControlFrame(ctx, websocketOpcodeClose, nil); err != nil {
			return nil, true, err
		}

		return nil, true, io.EOF
	default:
		return nil, true, fmt.Errorf("%w: %d", errUnsupportedWebSocketOpcode, opcode)
	}
}

func (c *httpConn) readMaskedPayload(payloadLength int) ([]byte, error) {
	var maskKey [4]byte
	if _, err := io.ReadFull(c.reader, maskKey[:]); err != nil {
		return nil, err
	}

	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return nil, err
	}

	for idx := range payload {
		payload[idx] ^= maskKey[idx%len(maskKey)]
	}

	return payload, nil
}

func writePayloadLength(writer *bufio.Writer, payloadLength int) error {
	switch {
	case payloadLength <= websocketPayloadInline:
		return writeInlinePayloadLength(writer, payloadLength)
	case payloadLength <= math.MaxUint16:
		if err := writer.WriteByte(websocketPayload126); err != nil {
			return err
		}

		return writeUint16(writer, uint16(payloadLength))
	default:
		if err := writer.WriteByte(websocketPayload127); err != nil {
			return err
		}

		return writeUint64(writer, uint64(payloadLength))
	}
}

func writeInlinePayloadLength(writer *bufio.Writer, payloadLength int) error {
	if payloadLength < 0 || payloadLength > websocketPayloadInline {
		return fmt.Errorf("%w: %d", errInlinePayloadLengthOutOfRange, payloadLength)
	}

	return writer.WriteByte(uint8(payloadLength))
}

func writeUint16(writer *bufio.Writer, value uint16) error {
	var rawLength [2]byte
	binary.BigEndian.PutUint16(rawLength[:], value)

	_, err := writer.Write(rawLength[:])

	return err
}

func writeUint64(writer *bufio.Writer, value uint64) error {
	var rawLength [8]byte
	binary.BigEndian.PutUint64(rawLength[:], value)

	_, err := writer.Write(rawLength[:])

	return err
}

func headerContainsToken(header http.Header, key, want string) bool {
	for _, value := range header.Values(key) {
		for token := range strings.SplitSeq(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), want) {
				return true
			}
		}
	}

	return false
}

func buildWebSocketAccept(secKey string) string {
	// #nosec G401 -- RFC 6455 requires SHA-1 for Sec-WebSocket-Accept.
	hash := sha1.Sum([]byte(secKey + websocketMagicGUID))

	return base64.StdEncoding.EncodeToString(hash[:])
}

func readWebSocketPayloadLength(reader *bufio.Reader, payloadLengthCode byte) (int, error) {
	switch payloadLengthCode {
	case websocketPayload126:
		var rawLength [2]byte
		if _, err := io.ReadFull(reader, rawLength[:]); err != nil {
			return 0, err
		}

		return int(binary.BigEndian.Uint16(rawLength[:])), nil
	case websocketPayload127:
		var rawLength [8]byte
		if _, err := io.ReadFull(reader, rawLength[:]); err != nil {
			return 0, err
		}

		payloadLength := binary.BigEndian.Uint64(rawLength[:])
		if payloadLength > uint64(maxWebSocketPayloadSize) {
			return 0, errWebSocketPayloadTooLarge
		}

		return int(payloadLength), nil
	default:
		return int(payloadLengthCode), nil
	}
}

func watchConnDeadline(ctx context.Context, setDeadline func(time.Time) error) func() {
	if deadline, ok := ctx.Deadline(); ok {
		ignoreDeadlineError(setDeadline(deadline))
	} else {
		ignoreDeadlineError(setDeadline(time.Time{}))
	}

	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			ignoreDeadlineError(setDeadline(time.Now()))
		case <-done:
		}
	}()

	return func() {
		close(done)
		ignoreDeadlineError(setDeadline(time.Time{}))
	}
}

func ignoreDeadlineError(err error) {
	if err != nil {
		return
	}
}
