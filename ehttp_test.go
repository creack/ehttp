package ehttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleFunc(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("fail")
	}
	HandleFunc("/", hdlr)

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, "fail", string(body))
}

func TestServeHTTP(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("fail")
	}
	http.Handle("/", HandlerFunc(hdlr))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, "fail", string(body))
}

func TestMWErrorPanicCommon(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	var (
		file string
		name string
		line int
	)
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		name, file, line = getCallstack(0)
		panic(fmt.Errorf("fail"))
	}
	http.HandleFunc("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
}

func TestMWErrorPanicEHTTP(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	var (
		file string
		name string
		line int
	)
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		name, file, line = getCallstack(0)
		panic(NewErrorf(http.StatusTeapot, "fail"))
	}
	http.HandleFunc("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusTeapot, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
}

func TestMWErrorPanicInt(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	var (
		file string
		name string
		line int
	)
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		name, file, line = getCallstack(0)
		panic(http.StatusTeapot)
	}
	http.HandleFunc("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, fmt.Sprintf("[%s %s:%d] (int) %d", name, file, line+1, http.StatusTeapot), string(body))
}

func TestMWErrorPanicMiddleware(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	var (
		file string
		name string
		line int
	)
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		name, file, line = getCallstack(0)
		panic(fmt.Errorf("fail"))
	}
	middleware := func(hdlr HandlerFunc) HandlerFunc {
		return HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
			return hdlr(w, req)
		})
	}
	http.HandleFunc("/", MWErrorPanic(middleware(hdlr)))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
}

func TestMWErrorPanicRuntimePanic(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	var (
		file string
		name string
		line int
	)
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		name, file, line = getCallstack(0)
		_ = (*http.Request)(nil).Body // Expected nil pointer dereference for test.
		return nil
	}
	middleware := func(hdlr HandlerFunc) HandlerFunc {
		return HandlerFunc(func(w http.ResponseWriter, req *http.Request) error {
			return hdlr(w, req)
		})
	}
	http.HandleFunc("/", MWErrorPanic(middleware(hdlr)))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertJSONError(t, fmt.Sprintf("[%s %s:%d] runtime error: invalid memory address or nil pointer dereference", name, file, line+1), string(body))
}

func TestDefaultMWError(t *testing.T) {
	// Use stdlib mux.
	mux := http.NewServeMux()

	// Register an ehttp.HandlerFunc via MWError use (*net/http.ServeMux).HandleFunc.
	mux.HandleFunc("/a", MWError(func(w http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("fail a")
	}))
	// Register an ehttp.HandlerFunc via MWError use (*net/http.ServeMux).Handle.
	mux.Handle("/b", MWError(func(w http.ResponseWriter, req *http.Request) error {
		return NewErrorf(http.StatusTeapot, "fail b")
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test /a.
	resp, err := http.Get(ts.URL + "/a")
	if err != nil {
		t.Fatalf("Error requesting test server /a: %s", err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	assertString(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	jsonErr := &JSONError{}
	err = json.NewDecoder(resp.Body).Decode(jsonErr)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Error parsing error output: %s", err)
	}
	assertInt(t, 1, len(jsonErr.Errors))
	assertString(t, "fail a", jsonErr.Errors[0])

	// Test /b.
	resp, err = http.Get(ts.URL + "/b")
	if err != nil {
		t.Fatalf("Error requesting test server /b: %s", err)
	}
	assertInt(t, http.StatusTeapot, resp.StatusCode)
	assertString(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	jsonErr = &JSONError{}
	err = json.NewDecoder(resp.Body).Decode(jsonErr)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Error parsing error output: %s", err)
	}
	assertInt(t, 1, len(jsonErr.Errors))
	assertString(t, "fail b", jsonErr.Errors[0])
}

func TestCustomErrorCallback(t *testing.T) {
	callback := func(w ResponseWriter, req *http.Request, err error) {
		fmt.Fprintf(w, "custom error callback: %s", err)
	}
	mux := NewServeMux(callback, "application/json", false, nil)

	// Register an ehttp.HandlerFunc via MWError use (*net/http.ServeMux).HandleFunc.
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("fail")
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error requesting test server: %s", err)
	}
	assertInt(t, http.StatusInternalServerError, resp.StatusCode)
	assertString(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("Error parsing error output: %s", err)
	}
	assertString(t, "custom error callback: fail", string(body))
}
