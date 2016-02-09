package ehttp

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Enforce that ResponseWriter implements the proper http interaces.
var (
	_ http.ResponseWriter = (*ResponseWriter)(nil)
	_ http.Hijacker       = (*ResponseWriter)(nil)
	_ http.Flusher        = (*ResponseWriter)(nil)
)

func TestResponseWriterHijack(t *testing.T) {
	r := &ResponseWriter{}
	if _, _, err := r.Hijack(); err == nil {
		t.Fatalf("Empty response writer should fail to hijack")
	}

	ts := httptest.NewServer(HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return fmt.Errorf("Error hijacking connection: %s", err)
		}
		fmt.Fprintf(conn, "hello")
		return conn.Close()
	}))
	defer ts.Close()

	client, err := net.Dial("tcp", strings.TrimPrefix(ts.URL, "http://"))
	if err != nil {
		t.Fatalf("Error trying to connect to the test server: %s", err)
	}
	fmt.Fprintf(client, "GET / HTTP/1.1\r\nConnection: Upgrade\r\n\r\n")
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
