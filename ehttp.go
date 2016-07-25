// Package ehttp wraps the standard http package with an
// error management middleware.
// Support net/http, gorilla/mux and any router using http.HandleFunc
package ehttp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
)

// ServeMux wraps *net/http.ServeMux for ehttp.
type ServeMux struct {
	ServeMux         *http.ServeMux
	errorContentType string                                     // Content-Type to use for the error cases.
	recoverPanic     bool                                       // Flag to know whether or not to recover from panics.
	log              *log.Logger                                // Custom logger to use for errors.
	sendError        func(ResponseWriter, *http.Request, error) // Callback to send error to the client.
}

// NewServeMux emulates net/http.NewServeMux but returns a *github.com/creack/ehttp.ServeMux.
// Logger default to the default log.Logger if nil.
// sendErrorCallback default to `fmt.Fprintf(w, "%s\n", err)` if nil.
func NewServeMux(sendErrorCallback func(ResponseWriter, *http.Request, error), errorContentType string, recoverPanic bool, logger *log.Logger) *ServeMux {
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if sendErrorCallback == nil {
		sendErrorCallback = func(w ResponseWriter, req *http.Request, err error) {
			fmt.Fprintf(w, "%s\n", err)
		}
	}
	return &ServeMux{
		ServeMux:         http.NewServeMux(),
		sendError:        sendErrorCallback,
		errorContentType: errorContentType,
		recoverPanic:     recoverPanic,
		log:              logger,
	}
}

// HandlerFunc converts the github.com/creack/ehttp.HandlerFunc handler to a standard http.Handler.
// If the recoverPanic flag is set, handle panics and return them as error.
func (sm *ServeMux) HandlerFunc(handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	if sm.recoverPanic {
		return sm.MWErrorPanic(handler)
	}
	return sm.MWError(handler)
}

// HandleFunc adds the given handler to the underlying ServeMux.
func (sm *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	sm.ServeMux.Handle(pattern, sm.HandlerFunc(handler))
}

// ServeHTTP implements http.Handler interface.
func (sm *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sm.ServeMux.ServeHTTP(w, req)
}

// MWError is the main middleware. When an error is returned, it send
// the data to the client if the header hasn't been sent yet, otherwise, log them.
func (sm *ServeMux) MWError(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ww := NewResponseWriter(w)
		if err := handler(ww, req); err != nil {
			sm.HandleError(ww, req, err)
			return
		}
	}
}

// MWErrorPanic wraps MWError and recovers from panic.
func (sm *ServeMux) MWErrorPanic(handler HandlerFunc) http.HandlerFunc {
	return sm.MWError(func(w http.ResponseWriter, req *http.Request) (err error) {
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
				err = sm.HandlePanic(err, e1)
				if e2, ok := err.(*Error); ok {
					e2.error = fmt.Errorf("[%s %s:%d] %s", name, file, line, e2.error)
				} else {
					err = fmt.Errorf("[%s %s:%d] %s", name, file, line, err)
				}
			}
		}()
		return handler(w, req)
	})
}

// HandleError handles the returned error from the MWError middleware.
// Should not be manually called. Exposed to be accessed from adaptor subpackages.
// If the error is nil, then no http code is yielded.
func (sm *ServeMux) HandleError(w ResponseWriter, req *http.Request, err error) {
	if code := w.Code(); code != 0 {
		sm.log.Printf("HTTP Error (header already sent): %s (%d)", err, code)
		return
	}
	if sm.errorContentType != "" {
		w.Header().Set("Content-Type", sm.errorContentType)
	}
	if err == nil {
		return
	}
	if e1, ok := err.(*Error); ok && e1.Code() != 0 {
		w.WriteHeader(e1.Code())
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	sm.sendError(w, req, err)
}

// HandlePanic handles the panic from the handler.
// Should not be manually called. Exposed to be accessed from adaptor subpackages.
func (sm *ServeMux) HandlePanic(err error, e1 interface{}) error {
	if e1 == nil { // no panic, return the given error.
		return err
	}
	if err != nil { // an error is already present. log the panic and return the given error.
		sm.log.Printf("Handler panic recovered: %v", e1)
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

// DefaultServeMux is the default ServeMux used by Serve.
// The behavior of the DefaultServeMux is as follows:
// - Logger:             Standard log.Logger.
// - Content-Type:       "application/json; charset=utf-8"
// - SendError callback: Send errors wraps in a json object with the key "errors" as a type []string.
// - RecoverPanic:       false.
var DefaultServeMux = &ServeMux{
	ServeMux: http.DefaultServeMux,
	sendError: func(w ResponseWriter, _ *http.Request, err error) {
		_ = json.NewEncoder(w).Encode(&JSONError{Errors: []string{err.Error()}})
	},
	errorContentType: "application/json; charset=utf-8",
	recoverPanic:     false,
	log:              log.New(os.Stderr, "", log.LstdFlags),
}

// HandlerFunc is the custom http handler function extending the standard one with
// an error return.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface, serving our custom prototype
// with error support.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ww := NewResponseWriter(w)
	if err := f(ww, req); err != nil {
		DefaultServeMux.HandleError(ww, req, err)
		return
	}
}

// JSONError is the default struct returned to the client upon error.
type JSONError struct {
	Errors []string `json:"errors"`
}

// HandleError exposes HandleError from the DefaultServeMux.
func HandleError(w ResponseWriter, req *http.Request, err error) {
	DefaultServeMux.HandleError(w, req, err)
}

// HandlePanic exposes HandlePanic from the DefaultServeMux.
func HandlePanic(err error, e1 interface{}) error {
	return DefaultServeMux.HandlePanic(err, e1)
}

// MWError exposes MWError from DefaultServeMux.
func MWError(handler HandlerFunc) http.HandlerFunc {
	return DefaultServeMux.MWError(handler)
}

// MWErrorPanic exposes MWErrorPanic from DefaultServeMux.
func MWErrorPanic(handler HandlerFunc) http.HandlerFunc {
	return DefaultServeMux.MWErrorPanic(handler)
}

// HandleFunc wraps net/http.HandleFunc with the error middleware.
func HandleFunc(pattern string, handler HandlerFunc) {
	DefaultServeMux.HandleFunc(pattern, handler)
}
