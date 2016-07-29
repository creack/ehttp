package ehttp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Enforce that *response and *http2responseWriter implements the proper interaces.
var (
	_ io.ReaderFrom       = (*response)(nil)
	_ io.Writer           = (*response)(nil)
	_ writeStringer       = (*response)(nil)
	_ http.CloseNotifier  = (*response)(nil)
	_ http.Flusher        = (*response)(nil)
	_ http.Hijacker       = (*response)(nil)
	_ http.ResponseWriter = (*response)(nil)

	_ io.Writer           = (*http2responseWriter)(nil)
	_ writeStringer       = (*http2responseWriter)(nil)
	_ http.CloseNotifier  = (*http2responseWriter)(nil)
	_ http.Flusher        = (*http2responseWriter)(nil)
	_ http.ResponseWriter = (*http2responseWriter)(nil)
)

func TestResponseWriterHijackNotHijcaker(t *testing.T) {
	respW := httptest.NewRecorder()
	r := &response{
		&http2responseWriter{
			ResponseWriter: respW,
		},
	}
	if _, _, err := r.Hijack(); err == nil {
		t.Fatalf("*net/http/httptest.ResonseRecoreded hijack should fail.")
	}

}

func TestResponseWriterHijack(t *testing.T) {
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return fmt.Errorf("Error hijacking connection: %s (%T)", err, w)
		}
		fmt.Fprintf(conn, "hello")
		return conn.Close()
	}))
	defer ts.Close()

	client, err := net.Dial("tcp", strings.TrimPrefix(ts.URL, "http://"))
	if err != nil {
		t.Fatalf("Error trying to connect to the test server: %s", err)
	}
	fmt.Fprintf(client, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\n\r\n")
	buf := make([]byte, 512)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}
}

func TestResponseWriterFlush(t *testing.T) {
	ch := make(chan int)
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprintf(w, "hello")
		w.(http.Flusher).Flush()
		<-ch
		return nil
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}
	close(ch)
}

func TestResponseWriterCloseNotifer(t *testing.T) {
	var ch <-chan bool
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		ch = w.(http.CloseNotifier).CloseNotify()
		// Disable keep alive.
		w.Header().Set("Connection", "Close")
		fmt.Fprintf(w, "hello")
		return nil
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}

	// Wait for the notification.
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-ch:
	case <-timer.C:
		t.Fatal("Timeout waiting for the CloseNotifier notification")
	}
}

func TestResponseWriterWriteString(t *testing.T) {
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		if _, ok := w.(writeStringer); !ok {
			return fmt.Errorf("responseWriter is not a writeString: %T", w)
		}
		_, err := io.WriteString(w, "hello")
		return err
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}
}

func TestResponseWriterReadFrom(t *testing.T) {
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		if _, ok := w.(io.ReaderFrom); !ok {
			return fmt.Errorf("responseWriter is not a readerFrom: (%T)", w)
		}
		rr, ww := io.Pipe()
		go func() {
			_, _ = io.WriteString(ww, "hello")
			_ = ww.Close()
		}()
		_, err := io.Copy(w, rr)
		return err
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}
}

func TestResponseWriterNotReadFrom(t *testing.T) {
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		// Change the underlying type to http2responseWriter which is not an io.ReaderFrom.
		w = &response{
			&http2responseWriter{
				ResponseWriter: &http2responseWriter{
					ResponseWriter: w,
				},
			},
		}
		if _, ok := w.(io.ReaderFrom); !ok {
			return fmt.Errorf("responseWriter is not a readerFrom: (%T)", w)
		}
		rr, ww := io.Pipe()
		go func() {
			_, _ = io.WriteString(ww, "hello")
			_ = ww.Close()
		}()
		_, err := io.Copy(w, rr)
		if err != ErrNotReaderFrom {
			return fmt.Errorf("Expected error when using io.ReaderFrom while not being implemented. Got: %v", err)
		}
		return nil
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		buf := bytes.NewBuffer(nil)
		_, _ = io.Copy(buf, resp.Body)
		t.Fatalf("Unexpected status code: %d (%s)", resp.StatusCode, buf)
	}
}

// dummyResponseWriter implements ehttp.ResponseWriter, but does not implement writeStringer.
type dummyResponseWriter struct{ http.ResponseWriter }

func (*dummyResponseWriter) Code() int { return 0 }

func TestResponseWriterNotWriteString(t *testing.T) {
	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		w = &response{
			&http2responseWriter{
				ResponseWriter: dummyResponseWriter{ResponseWriter: w},
				code:           new(int32),
			},
		}
		// Event though not implemented, it should work using io.Writer.
		_, err := io.WriteString(w, "hello")
		return err
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error fetching test server: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Error reading from test server: %s", err)
	}
	buf = buf[:n]
	if expect, got := "hello", string(buf); expect != got {
		t.Fatalf("Unexpected message from test server.\nExpect:\t%s\nGot:\t%s", expect, got)
	}
}

func TestNewResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := NewResponseWriter(recorder)
	w1 := NewResponseWriter(recorder)
	w2 := NewResponseWriter(w)

	if w != w2 {
		t.Fatalf("NewResposneWriter called with an ehttp.ResopnseWriter should return the given writer")
	}
	if w == w1 {
		t.Fatalf("NewResposneWriter called with a regular http.ResponseWriter should wrap it and return a new object")
	}
}
