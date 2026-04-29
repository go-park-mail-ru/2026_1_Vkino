package metrics

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

func InstrumentHTTPHandler(service, route string, next http.Handler) http.Handler {
	Register()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		mw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(mw, r)

		status := strconv.Itoa(mw.StatusCode())
		serviceLabel := labelValue(service, "unknown")
		routeLabel := labelValue(route, "unknown")

		HTTPRequestsTotal.WithLabelValues(serviceLabel, r.Method, routeLabel, status).Inc()
		HTTPRequestDurationSeconds.WithLabelValues(serviceLabel, r.Method, routeLabel, status).
			Observe(time.Since(startedAt).Seconds())

		if mw.StatusCode() >= http.StatusInternalServerError {
			HTTPRequestErrorsTotal.WithLabelValues(serviceLabel, r.Method, routeLabel, status).Inc()
		}
	})
}

func InstrumentHTTPHandlerFunc(service, route string, next http.HandlerFunc) http.HandlerFunc {
	return InstrumentHTTPHandler(service, route, next).ServeHTTP
}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *responseWriter) StatusCode() int {
	return w.statusCode
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(p)
}

func (w *responseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	flusher.Flush()
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not implement http.Hijacker")
	}

	return hijacker.Hijack()
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}

	return pusher.Push(target, opts)
}

func (w *responseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	readerFrom, ok := w.ResponseWriter.(io.ReaderFrom)
	if ok {
		return readerFrom.ReadFrom(r)
	}

	return io.Copy(w.ResponseWriter, r)
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
