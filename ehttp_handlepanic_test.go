package ehttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandlePanicNil(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	if err := HandlePanic(nil, nil); err != nil {
		t.Fatal(err)
	}
	assertInt(t, 0, buf.Len())
}

func TestHandlePanicError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	err := HandlePanic(nil, fmt.Errorf("fail"))
	assertInt(t, 0, buf.Len())
	assertString(t, "fail", err.Error())
}

func TestHandlePanicIncomingError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	e1 := fmt.Errorf("fail")
	err := HandlePanic(e1, nil)
	if err != e1 {
		t.Errorf("error passed to HandlePanic does not match returned one when no panic")
	}
	assertInt(t, 0, buf.Len())
	assertString(t, "fail", err.Error())
}

func TestHandlePanicBothError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	e1 := fmt.Errorf("hello")
	err := HandlePanic(e1, fmt.Errorf("fail"))
	if err != e1 {
		t.Errorf("error passed to HandlePanic does not match returned one when error already present")
	}
	if !strings.Contains(buf.String(), "panic recovered: fail") {
		t.Errorf("Panic error value not found in log output.\nGot: %s", buf.String())
	}
	assertString(t, "hello", err.Error())
}

func TestHandlePanicNonError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	err := HandlePanic(nil, "fail")
	assertInt(t, 0, buf.Len())
	assertString(t, "(string) fail", err.Error())
}

func TestFullHandlePanicError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	mux := NewServeMux(nil, "", true, log.New(buf, "", log.LstdFlags))

	mux.HandleFunc("/", func(http.ResponseWriter, *http.Request) error {
		panic(NewErrorf(http.StatusTeapot, "fail"))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error sending http request to test server: %s", err)
	}
	assertInt(t, http.StatusTeapot, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Error reading body from test server request: %s", err)
	}

	assertInt(t, 0, buf.Len())
	if !strings.Contains(string(body), "fail") {
		t.Fatalf("Unexpected response body from panic. Expected to see %q, got: %q", "fail", body)
	}
}
