package ehttp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
)

// TODO: add support for *net/http/httptest.ResponseRecorder. (Missing net/http.CloseNotifer interface).

// Common errors.
var (
	ErrNotHijacker   = errors.New("not a net/http.Hijacker")
	ErrNotReaderFrom = errors.New("not an io.ReaderFrom")
)

// From io.stringWriter.
//
// NOTE: Why is this not exposed in the stdlib? Looks like it is reimplemented a lot:
// - io.stringWriter
// - strings.stringWriterIface
// - net/http.writeStringer
// - net/http.http2stringWriter
// - github.com/creack/ehttp.writeStringer
type writeStringer interface {
	WriteString(string) (int, error)
}

// ResponseWriter extends http.ResponseWriter and exposes the http status code.
type ResponseWriter interface {
	http.ResponseWriter
	Code() int
}

// http2responseWriter implements ehttp.ResponseWriter and exposes all *net/http.http2responseWriter interfaces.
// stores the http status code as it gets se/sent.
// As go go1.6.3:
// - pointer type *net/http.http2responseWriter
//   - implements io.Writer
//   - implements io.stringWriter
//   - implements CloseNotifier
//   - implements Flusher
//   - implements ResponseWriter
//   - implements http2stringWriter
//   - implements writeStringer
//   - implements strings.stringWriterIface
type http2responseWriter struct {
	http.ResponseWriter
	code *int32
}

// response implements ehttp.ResponseWriter and exposes all *net/http.response interfaces.
// stores the http status code as it gets set/sent.
// It has the same intefaces as *net/http.http2responseWriter plus net/http.Hijacker and io.ReaderFrom.
// As of go1.6.3:
// - pointer type *net/http.response
//   - implements io.ReaderFrom
//   - implements io.Writer
//   - implements io.stringWriter
//   - implements net/http.CloseNotifier
//   - implements net/http.Flusher
//   - implements net/http.Hijacker
//   - implements net/http.ResponseWriter
//   - implements net/http.http2stringWriter
//   - implements net/http.writeStringer
//   - implements strings.stringWriterIface
type response struct {
	*http2responseWriter
}

// NewResponseWriter instantiates a new ehttp ResponseWriter.
//
// If the underlying http.ResponseWriter implement net/http.Hijacker,
// assume it is a *net/http.response and use *github.com/creack/ehttp.response, otherwise,
// assume it is a *net/http.http2responseWriter and use *github.com/creack/ehttp.http2responseWriter.
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	// If w is already an ehttp.ReponseWriter, return it.
	if ww, ok := w.(ResponseWriter); ok {
		return ww
	}
	ret := &http2responseWriter{
		ResponseWriter: w,
		code:           new(int32),
	}
	if _, ok := w.(http.Hijacker); ok {
		return &response{
			http2responseWriter: ret,
		}
	}
	return ret
}

// Code return the http code stored in the response writer.
func (w http2responseWriter) Code() int {
	return int(atomic.LoadInt32(w.code))
}

// WriteHeader wraps underlying WriteHeader.
// - sends the header only if not already sent.
// - flag that the headers have been sent
// - store the sent code
func (w *http2responseWriter) WriteHeader(code int) {
	if atomic.CompareAndSwapInt32(w.code, 0, int32(code)) {
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write wraps the underlying Write and flag that the headers have been sent.
// Following the net/http behavior, if no http status code has been set, assume http.StatusOK.
func (w *http2responseWriter) Write(buf []byte) (int, error) {
	atomic.CompareAndSwapInt32(w.code, 0, int32(http.StatusOK))
	return w.ResponseWriter.Write(buf)
}

// WriteString wraps the underlying WriteString if available. Use ehttp.response.Write if not.
// Flag the headers as sent.
func (w *http2responseWriter) WriteString(str string) (int, error) {
	atomic.CompareAndSwapInt32(w.code, 0, int32(http.StatusOK))
	if ws, ok := w.ResponseWriter.(writeStringer); ok {
		return ws.WriteString(str)
	}
	return w.Write([]byte(str))
}

// Flush exposes the underlying net/http.Flusher interface. No-op if not a Flusher.
func (w *http2responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// CloseNotify exposes the underlying net/http.CloseNotifier interface.
// Panics if not available.
func (w *http2responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Hijack exposes the underlying net/http.Hijacker if available.
// Errors out if not.
func (w *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, ErrNotHijacker
}

// ReadFrom exposes the underlying io.ReaderFrom interface if available.
// Errors out if not.
func (w *response) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return 0, ErrNotReaderFrom
}
