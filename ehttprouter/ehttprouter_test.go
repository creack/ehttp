package ehttprouter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/creack/ehttp"
	"github.com/julienschmidt/httprouter"
)

func TestMWErrorCommon(t *testing.T) {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		return fmt.Errorf("fail")
	}
	router := httprouter.New()
	router.GET("/", MWError(hdlr))

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
	assertJSONError(t, "fail", string(body))
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
	assertJSONError(t, "fail", string(body))
}

func TestMWErrorPanicEHTTP(t *testing.T) {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(ehttp.NewErrorf(http.StatusTeapot, "fail"))
	}
	router := httprouter.New()
	router.GET("/", MWErrorPanic(hdlr))

	ts := httptest.NewServer(router)
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
	assertJSONError(t, "fail", string(body))
}

func TestMWErrorPanicInt(t *testing.T) {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(http.StatusTeapot)
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
	assertJSONError(t, fmt.Sprintf("(int) %d", http.StatusTeapot), string(body))
}

func TestWrappedHelperMethods(t *testing.T) {
	router := New(nil, "", true, nil)

	testHandler := func(w http.ResponseWriter, req *http.Request, params httprouter.Params) error {
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	}

	methods := map[string]func(string, Handle){
		"GET":     router.GET,
		"DELETE":  router.DELETE,
		"OPTIONS": router.OPTIONS,
		"PATCH":   router.PATCH,
		"POST":    router.POST,
		"PUT":     router.PUT,
	}
	for _, method := range methods {
		method("/", testHandler)
	}

	ts := httptest.NewServer(router)
	defer ts.Close()

	for methodName := range methods {
		req, err := http.NewRequest(methodName, ts.URL, nil)
		if err != nil {
			t.Fatalf("Error creating new request for method %s: %s", methodName, err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error connecting to test server for method %s: %s", methodName, err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("Error reading response from test server for method %s: %s", methodName, err)
		}
		assertInt(t, http.StatusTeapot, resp.StatusCode)
		assertString(t, "fail", string(body))
		if t.Failed() {
			t.Logf("Assert error for method %s", methodName)
			break
		}
	}

	// Specicif test for HEAD.
	router = New(nil, "", false, nil)
	ch := make(chan struct{}, 1)
	defer close(ch)
	testHandler = func(w http.ResponseWriter, req *http.Request, params httprouter.Params) error {
		ch <- struct{}{}
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	}
	router.HEAD("/", testHandler)
	ts2 := httptest.NewServer(router)
	defer ts2.Close()

	resp, err := http.Head(ts2.URL)
	if err != nil {
		t.Fatalf("Error connecting to test server for method %s: %s", "HEAD", err)
	}
	assertInt(t, http.StatusTeapot, resp.StatusCode)

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		t.Fatal("Timeout waiting fro the HEAD call")
	case <-ch:
	}
}

func TestDummyPanicHandler(t *testing.T) {
	router := New(nil, "", false, nil)
	// No-op. Makeing sure that it is a function and not a field.
	router.PanicHandler()
}

func TestWrappedHandleMethods(t *testing.T) {
	testHandler := func(w http.ResponseWriter, req *http.Request, params httprouter.Params) error {
		return ehttp.BadRequest
	}

	router := New(nil, "", false, nil)
	router.Handle("GET", "/a", testHandler)
	router.HandlerFunc("GET", "/b", func(http.ResponseWriter, *http.Request) {})
	router.Handler("GET", "/c", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/a")
	if err != nil {
		t.Fatalf("Error connecting to test server: %s", err)
	}
	_ = resp.Body.Close()
	assertInt(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/b")
	if err != nil {
		t.Fatalf("Error connecting to test server: %s", err)
	}
	_ = resp.Body.Close()
	assertInt(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/c")
	if err != nil {
		t.Fatalf("Error connecting to test server: %s", err)
	}
	_ = resp.Body.Close()
	assertInt(t, http.StatusConflict, resp.StatusCode)
}

func TestWrappedSpecialMethods(t *testing.T) {
	router := New(nil, "", false, nil)
	testHandler := func(http.ResponseWriter, *http.Request, httprouter.Params) error {
		return nil
	}
	router.GET("/b", testHandler)

	if handle, _, _ := router.Lookup("GET", "/a"); handle != nil {
		t.Fatal("Unset handler path lookup should return a nil handler")
	}
	if handle, _, _ := router.Lookup("GET", "/b"); handle == nil {
		t.Fatal("Set handler path lookup should return the handler")
	}

	// Dummy call for coverage. Already tested in httprouter package.
	router.ServeFiles("/f/*filepath", nil)
}
