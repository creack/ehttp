// Package ehttp wraps the standard http package with an
// error management middleware.
// Support net/http, gorilla/mux and any router using http.HandleFunc
package ehttp

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
)

// HandlerFunc is a custom http handler extending the standard one with
// an error return.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// HandleError handles the returned error from the MWError middleware.
// Should not be manually called. Exposed to be accessed from adaptor subpackages.
func HandleError(w *ResponseWriter, err error) {
	if code := w.Code(); code != 0 {
		log.Printf("HTTP Error (header already sent): %s (%d)", err, code)
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
		return fmt.Errorf("(%T) %v", e1, e1)
	}
	// return the panic error
	return e2
}

func (hdlr HandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ww := NewResponseWriter(w)
	if err := hdlr(ww, req); err != nil {
		HandleError(ww, err)
		return
	}
}

// MWError is the main middleware. When an error is returned, it send
// the data to the client if the header hasn't been sent yet, otherwise, log them.
func MWError(hdlr HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ww := NewResponseWriter(w)
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
				var name string
				skip := 0
			begin:
				pc, file, line, ok := runtime.Caller(3 + skip)
				if ok {
					name = runtime.FuncForPC(pc).Name()
					name = path.Base(name)
					file = path.Base(file)
					// If there is a runtime panic (nil dereference or other) we endup with a runtime callstack. Skip until out of it.
					if len(name) > 7 && name[:7] == "runtime" {
						skip++
						goto begin
					}
				}
				err = HandlePanic(err, e1)
				if e2, ok := err.(*Error); ok {
					e2.error = fmt.Errorf("[%s %s:%d] %s", name, file, line, e2.error)
				} else {
					err = fmt.Errorf("[%s %s:%d] %s", name, file, line, err)
				}
			}
		}()
		return hdlr(w, req)
	})
}

// HandleFunc wraps http.handleFunc with the error middleware.
func HandleFunc(path string, handler HandlerFunc) {
	http.HandleFunc(path, MWError(handler))
}
