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

// WriteHeader wraps underlying WriteHeader
// - flag that the headers have been sent
// - store the sent code
func (w *ResponseWriter) WriteHeader(code int) {
	atomic.CompareAndSwapInt32(w.code, 0, int32(code))
	w.ResponseWriter.WriteHeader(code)
}

// Write wraps the underlying Write and flag that the headers have been sent.
func (w *ResponseWriter) Write(buf []byte) (int, error) {
	atomic.CompareAndSwapInt32(w.code, 0, int32(http.StatusOK))
	return w.ResponseWriter.Write(buf)
}
