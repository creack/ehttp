// Package ehttp wraps the standard http package with an
// error management middleware.
// Support net/http, gorilla/mux and any router using http.HandleFunc
package ehttp

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// HandlerFunc is a custom http handler extending the standard one with
// an error return.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// HandleError handles the returned error from the MWError middleware.
// Should not be manually called. Exposed to be accessed from adaptor subpackages.
func HandleError(w *ResponseWriter, err error) {
	if atomic.LoadInt32(&w.headerSent) != 0 {
		log.Printf("HTTP Error (header already sent): %s (%d)", err, atomic.LoadInt32(&w.code))
		return
	}
	if e1, ok := err.(*Error); ok {
		w.WriteHeader(e1.Code())
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintln(w, err)
}

// HandlePanic handles the panic from the handler.
// Should not be manually called. Exposed to be accessed from adaptor subpackages.
func HandlePanic(err error, e1 interface{}) error {
	if e1 == nil { // no panic, return the given error.
		return err
	}
	if err != nil { // an error is already present. log the panic and return the given error.
		log.Printf("Handler panic recovered: %v", e1)
		return err
	}
	// we have a panic and no given error.
	e2, ok := e1.(error)
	if !ok { // the panic is not an error, create an error out of the string representation of the panic.
		return fmt.Errorf("%v", e1)
	}
	// return the panic error
	return e2
}

func (hdlr HandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ww := &ResponseWriter{ResponseWriter: w}
	if err := hdlr(w, req); err != nil {
		HandleError(ww, err)
		return
	}
}

// MWError is the main middleware. When an error is returned, it send
// the data to the client if the header hasn't been sent yet, otherwise, log them.
func MWError(hdlr HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ww := &ResponseWriter{ResponseWriter: w}
		if err := hdlr(ww, req); err != nil {
			HandleError(ww, err)
			return
		}
	}
}

// MWErrorPanic wraps MWError and recovers from panic.
func MWErrorPanic(hdlr HandlerFunc) http.HandlerFunc {
	return MWError(func(w http.ResponseWriter, req *http.Request) (err error) {
		defer func() {
			if e1 := recover(); e1 != nil {
				err = HandlePanic(err, e1)
			}
		}()
		return hdlr(w, req)
	})
}

// HandleFunc wraps http.handleFunc with the error middleware.
func HandleFunc(path string, handler HandlerFunc) {
	http.HandleFunc(path, MWError(handler))
}
