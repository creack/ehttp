package ehttp

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandleErrorNil(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := &ResponseWriter{ResponseWriter: rec}
	HandleError(w, nil)
	assertInt(t, http.StatusInternalServerError, int(w.code))
	assertInt(t, 0, buf.Len())
	assertString(t, "<nil>\n", rec.Body.String())
}

func TestHandleErrorCommon(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := &ResponseWriter{ResponseWriter: rec}
	HandleError(w, fmt.Errorf("fail"))
	assertInt(t, http.StatusInternalServerError, int(w.code))
	assertInt(t, 0, buf.Len())
	assertString(t, "fail\n", rec.Body.String())
}

func TestHandleErrorEHTTP(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := &ResponseWriter{ResponseWriter: rec}
	HandleError(w, NewErrorf(418, "fail"))
	assertInt(t, 418, int(w.code))
	assertInt(t, 0, buf.Len())
	assertString(t, "fail\n", rec.Body.String())
}

func TestHandleErrorSentHeader(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	rec := httptest.NewRecorder()
	w := &ResponseWriter{ResponseWriter: rec}
	w.WriteHeader(http.StatusBadGateway)
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	HandleError(w, NewErrorf(418, "fail"))
	assertInt(t, http.StatusBadGateway, int(w.code))
	if !strings.Contains(buf.String(), fmt.Sprintf("%s (%d)", "fail", http.StatusBadGateway)) {
		t.Errorf("Error and status code not found in log output.\nGot: %s", buf.String())
	}
	assertString(t, "hello", rec.Body.String())
}
