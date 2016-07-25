package ehttp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleErrorNil(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := NewResponseWriter(rec)
	HandleError(w, nil, nil)
	assertInt(t, 0, w.Code())
	assertInt(t, 0, buf.Len())
	assertString(t, "", rec.Body.String())
}

func TestHandleErrorCommon(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := NewResponseWriter(rec)
	HandleError(w, nil, fmt.Errorf("fail"))
	assertInt(t, http.StatusInternalServerError, w.Code())
	assertInt(t, 0, buf.Len())
	assertJSONError(t, "fail", rec.Body.String())
}

func TestHandleErrorEHTTP(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := NewResponseWriter(rec)
	HandleError(w, nil, NewErrorf(http.StatusTeapot, "fail"))
	assertInt(t, http.StatusTeapot, w.Code())
	assertInt(t, 0, buf.Len())
	assertJSONError(t, "fail", rec.Body.String())
}

func TestHandleErrorSentHeader(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	DefaultServeMux.log.SetOutput(buf)
	defer DefaultServeMux.log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := NewResponseWriter(rec)
	w.WriteHeader(http.StatusBadGateway)
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	HandleError(w, nil, NewErrorf(http.StatusTeapot, "fail"))
	assertInt(t, http.StatusBadGateway, w.Code())
	if !strings.Contains(buf.String(), fmt.Sprintf("%s (%d)", "fail", http.StatusBadGateway)) {
		t.Errorf("Error and status code not found in log output.\nGot: %s", buf.String())
	}
	assertString(t, "hello", rec.Body.String())
}
