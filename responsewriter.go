package ehttp

import (
	"net/http"
	"sync/atomic"
)

// ResponseWriter wraps http.ResponseWriter and holds
// a flag to know if the headers have been sent as well has
// the http code sent.
type ResponseWriter struct {
	http.ResponseWriter
	headerSent int32
	code       int32
}

// WriteHeader wraps underlying WriteHeader
// - flag that the headers have been sent
// - store the sent code
func (w *ResponseWriter) WriteHeader(code int) {
	atomic.StoreInt32(&w.headerSent, 1)
	atomic.CompareAndSwapInt32(&w.code, 0, int32(code))
	w.ResponseWriter.WriteHeader(code)
}

// Write wraps the underlying Write and flag that the headers have been sent.
func (w *ResponseWriter) Write(buf []byte) (int, error) {
	atomic.StoreInt32(&w.headerSent, 1)
	return w.ResponseWriter.Write(buf)
}
