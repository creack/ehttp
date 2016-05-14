package ehttp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestError(t *testing.T) {
	err := Error{}
	assertString(t, "<nil>", err.Error())

	err = Error{code: http.StatusTeapot, error: fmt.Errorf("fail")}
	assertString(t, "fail", err.Error())

}

func TestNewError(t *testing.T) {
	err := NewError(http.StatusTeapot, io.EOF)
	assertString(t, "EOF", err.Error())

	err = NewErrorf(http.StatusTeapot, "hello %s", "world")
	assertString(t, "hello world", err.Error())
}

func TestGetError(t *testing.T) {
	var opError error = &net.OpError{Op: "op", Err: errors.New("fail")}

	e1 := NewError(http.StatusTeapot, opError)
	e2 := e1.(*Error).GetError()

	if expect, got := fmt.Sprintf("%v (%T)", opError, opError), fmt.Sprintf("%v (%T)", e2, e2); expect != got {
		t.Fatalf("Unexpected error returned by GetError().\nExpect:\t%s\nGot:\t%s\n", expect, got)
	}

}
