package ehttp

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestHandlePanicNil(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	if err := HandlePanic(nil, nil); err != nil {
		t.Fatal(err)
	}
	assertInt(t, 0, buf.Len())
}

func TestHandlePanicError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	err := HandlePanic(nil, fmt.Errorf("fail"))
	assertInt(t, 0, buf.Len())
	assertString(t, "fail", err.Error())
}

func TestHandlePanicIncomingError(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

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
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

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
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	err := HandlePanic(nil, "fail")
	assertInt(t, 0, buf.Len())
	assertString(t, "fail", err.Error())
}
