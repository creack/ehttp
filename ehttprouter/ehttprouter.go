// Package ehttprouter wraps httprouter with an
// error management middleware.
package ehttprouter

import (
	"net/http"

	"github.com/creack/ehttp"
	"github.com/julienschmidt/httprouter"
)

// Handle wraps httprouter's Handle and extends it with an error return.
type Handle func(http.ResponseWriter, *http.Request, httprouter.Params) error

// MWError is the middleware taking care of the returned error.
func MWError(hdlr Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		ww := &ehttp.ResponseWriter{ResponseWriter: w}
		if err := hdlr(ww, req, p); err != nil {
			ehttp.HandleError(ww, err)
			return
		}
	}
}

// MWErrorPanic wraps MWError and recovers from panic.
func MWErrorPanic(hdlr Handle) httprouter.Handle {
	return MWError(func(w http.ResponseWriter, req *http.Request, p httprouter.Params) (err error) {
		defer func() {
			if e1 := recover(); e1 != nil {
				err = ehttp.HandlePanic(err, e1)
			}
		}()
		return hdlr(w, req, p)
	})
}
