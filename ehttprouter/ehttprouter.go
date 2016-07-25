// Package ehttprouter wraps httprouter with an
// error management middleware.
package ehttprouter

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/julienschmidt/httprouter"
)

// Handle wraps httprouter's Handle and extends it with an error return.
type Handle func(http.ResponseWriter, *http.Request, httprouter.Params) error

// Router wraps *github.com/julienschmidt/httprouter.Router with error management.
type Router struct {
	*httprouter.Router                 // Underlying httprouter.Router.
	mux                *ehttp.ServeMux // ehttp mux.
	recoverPanic       bool            // Flag to know whether or not handle panics.
}

// DefaultRouter is the default router for direct access.
var DefaultRouter = &Router{
	Router:       httprouter.New(),
	mux:          ehttp.DefaultServeMux,
	recoverPanic: false,
}

// New instantiates a new ehttprouter.Router.
func New(sendErrorCallback func(ehttp.ResponseWriter, *http.Request, error), errorContentType string, recoverPanic bool, logger *log.Logger) *Router {
	return &Router{
		Router:       httprouter.New(),
		mux:          ehttp.NewServeMux(sendErrorCallback, errorContentType, recoverPanic, logger),
		recoverPanic: recoverPanic,
	}
}

// PanicHandler is a place holder to disable unwanted access to the underlying field.
// Panic is handled via ehttprouter instead.
func (r *Router) PanicHandler() {}

// middlewareSelect applies the error middleware.
// If the recoverPanic flag is set, recover panics, otherwise, just handle errors.
func (r *Router) middlewareSelect(handle Handle) httprouter.Handle {
	if r.recoverPanic {
		return r.MWErrorPanic(handle)
	}
	return r.MWError(handle)
}

// MWError is the middleware taking care of the returned error.
func (r *Router) MWError(handle func(http.ResponseWriter, *http.Request, httprouter.Params) error) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		ww := ehttp.NewResponseWriter(w)
		if err := handle(ww, req, p); err != nil {
			r.mux.HandleError(ww, req, err)
			return
		}
	}
}

// MWErrorPanic wraps MWError and recovers from panic.
func (r *Router) MWErrorPanic(handle Handle) httprouter.Handle {
	return r.MWError(func(w http.ResponseWriter, req *http.Request, p httprouter.Params) (err error) {
		defer func() {
			if e1 := recover(); e1 != nil {
				err = r.mux.HandlePanic(err, e1)
			}
		}()
		return handle(w, req, p)
	})
}

// MWError exposes the default router MWError method.
func MWError(handle Handle) httprouter.Handle {
	return DefaultRouter.MWError(handle)
}

// MWErrorPanic exposes the default router MWErrorPanic method.
func MWErrorPanic(handle Handle) httprouter.Handle {
	return DefaultRouter.MWErrorPanic(handle)
}
