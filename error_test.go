package ehttp

import (
	"fmt"
	"io"
	"testing"
)

func TestError(t *testing.T) {
	err := Error{}
	assertString(t, "<nil>", err.Error())

	err = Error{code: 418, error: fmt.Errorf("fail")}
	assertString(t, "fail", err.Error())

}

func TestNewError(t *testing.T) {
	err := NewError(418, io.EOF)
	assertString(t, "EOF", err.Error())

	err = NewErrorf(418, "hello %s", "world")
	assertString(t, "hello world", err.Error())
}
