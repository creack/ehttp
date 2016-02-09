package ehttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"strings"
	"testing"
)

func assertInt(t *testing.T, expect, got int) {
	_, file, line := getCallstack(1)
	if expect != got {
		t.Errorf("[%s:%d] Unexpected result.\nExpect:\t%d\nGot:\t%d\n", file, line, expect, got)
	}
}

func assertString(t *testing.T, expect, got string) {
	_, file, line := getCallstack(1)
	expect, got = strings.TrimSpace(expect), strings.TrimSpace(got)
	if expect != got {
		t.Errorf("[%s:%d] Unexpected result.\nExpect:\t%s\nGot:\t%s\n", file, line, expect, got)
	}
}

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
	assertString(t, "fail", string(body))
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
	assertString(t, "fail", string(body))
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
	assertString(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
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
		panic(NewErrorf(418, "fail"))
	}
	http.HandleFunc("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertInt(t, 418, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assertString(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
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
		panic(418)
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
	assertString(t, fmt.Sprintf("[%s %s:%d] (int) 418", name, file, line+1), string(body))
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
	assertString(t, fmt.Sprintf("[%s %s:%d] fail", name, file, line+1), string(body))
}

func getCallstack(skip int) (string, string, int) {
	var name string
	pc, file, line, ok := runtime.Caller(1 + skip)
	if !ok {
		name, file, line = "<unkown>", "<unknown>", -1
	} else {
		name = runtime.FuncForPC(pc).Name()
		name = path.Base(name)
		file = path.Base(file)
	}
	return name, file, line
}
