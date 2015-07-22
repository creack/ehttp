package ehttprouter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/creack/ehttp"
	"github.com/julienschmidt/httprouter"
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

func TestMWErrorPanicCommon(t *testing.T) {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(fmt.Errorf("fail"))
	}
	router := httprouter.New()
	router.GET("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(router)
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
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(ehttp.NewErrorf(418, "fail"))
	}
	router := httprouter.New()
	router.GET("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(router)
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
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(418)
	}
	router := httprouter.New()
	router.GET("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(router)
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
