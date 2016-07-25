package ehttp

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
)

// ResponseWriter wraps http.ResponseWriter and holds
// a flag to know if the headers have been sent as well has
// the http code sent.
type ResponseWriter struct {
	http.ResponseWriter
	code *int32
}

// NewResponseWriter instantiates a new ehttp ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		code:           new(int32),
	}
}

// Code return the http code stored in the response writer.
func (w ResponseWriter) Code() int {
	return int(atomic.LoadInt32(w.code))
}

// WriteHeader wraps underlying WriteHeader.
// - sends the header only if not already sent.
// - flag that the headers have been sent
// - store the sent code
func (w *ResponseWriter) WriteHeader(code int) {
	if atomic.CompareAndSwapInt32(w.code, 0, int32(code)) {
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write wraps the underlying Write and flag that the headers have been sent.
// Following the net/http behavior, if no http status code has been set, assume http.StatusOK.
func (w *ResponseWriter) Write(buf []byte) (int, error) {
	atomic.CompareAndSwapInt32(w.code, 0, int32(http.StatusOK))
	return w.ResponseWriter.Write(buf)
}

// Hijack wraps the underlying Hijack if available.
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("not a hijacker")
	}
	return hijacker.Hijack()
}

// Flush wraps the underlying Flush. Panic if not a Flusher.
func (w *ResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}
