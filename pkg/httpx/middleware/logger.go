package middleware

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
)

const requestIDHeader = "X-Request-ID"

var errResponseWriterHijacker = errors.New("response writer does not implement http.Hijacker")

func LoggerMiddleware(baseLogger *logger.Logger) func(http.Handler) http.Handler {
	if baseLogger == nil {
		baseLogger = logger.FromContext(context.TODO())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
			if requestID == "" {
				requestID = newRequestID()
			}

			requestLogger := baseLogger.WithField("request_id", requestID)
			ctx := logger.ContextWithLogger(r.Context(), requestLogger)

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			lw.Header().Set(requestIDHeader, requestID)

			startedAt := time.Now()

			next.ServeHTTP(lw, r.WithContext(ctx))

			requestLogger.
				WithField("method", r.Method).
				WithField("path", r.URL.Path).
				WithField("status", lw.StatusCode()).
				WithField("duration", time.Since(startedAt).String()).
				Info("request handled")
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter

	statusCode  int
	wroteHeader bool
}

func (w *loggingResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(p)
}

func (w *loggingResponseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	flusher.Flush()
}

func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errResponseWriterHijacker
	}

	return hijacker.Hijack()
}

func (w *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}

	return pusher.Push(target, opts)
}

func (w *loggingResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	readerFrom, ok := w.ResponseWriter.(io.ReaderFrom)
	if ok {
		return readerFrom.ReadFrom(r)
	}

	return io.Copy(w.ResponseWriter, r)
}

func (w *loggingResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func newRequestID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(id[:])
}
