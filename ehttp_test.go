package ehttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertInt(t *testing.T, expect, got int) {
	if expect != got {
		t.Errorf("Unexpected result.\nExpect:\t%d\nGot:\t%d\n", expect, got)
	}
}

func assertString(t *testing.T, expect, got string) {
	expect, got = strings.TrimSpace(expect), strings.TrimSpace(got)
	if expect != got {
		t.Errorf("Unexpected result.\nExpect:\t%s\nGot:\t%s\n", expect, got)
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
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
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
	assertString(t, "fail", string(body))
}

func TestMWErrorPanicEHTTP(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
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
	assertString(t, "fail", string(body))
}

func TestMWErrorPanicInt(t *testing.T) {
	defer func() { http.DefaultServeMux = http.NewServeMux() }()
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
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
	assertString(t, "418", string(body))
}
