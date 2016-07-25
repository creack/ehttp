package ehttprouter

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ServeHTTP exposes the underlying router's http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}

// Handle wraps the httprouter Handle.
func (r *Router) Handle(method, path string, handle Handle) {
	r.Router.Handle(method, path, r.middlewareSelect(handle))
}

// DELETE wraps underlying method.
func (r *Router) DELETE(path string, handle Handle) { r.Router.DELETE(path, r.middlewareSelect(handle)) }

// GET wraps underlying method.
func (r *Router) GET(path string, handle Handle) { r.Router.GET(path, r.middlewareSelect(handle)) }

// HEAD wraps underlying method.
func (r *Router) HEAD(path string, handle Handle) { r.Router.HEAD(path, r.middlewareSelect(handle)) }

// OPTIONS wraps underlying method.
func (r *Router) OPTIONS(path string, handle Handle) {
	r.Router.OPTIONS(path, r.middlewareSelect(handle))
}

// PATCH wraps underlying method.
func (r *Router) PATCH(path string, handle Handle) { r.Router.PATCH(path, r.middlewareSelect(handle)) }

// POST wraps underlying method.
func (r *Router) POST(path string, handle Handle) { r.Router.POST(path, r.middlewareSelect(handle)) }

// PUT wraps underlying method.
func (r *Router) PUT(path string, handle Handle) { r.Router.PUT(path, r.middlewareSelect(handle)) }

// Handler exposes the httprouter Handler method.
// NOTE: does not handle erors nor panics.
func (r *Router) Handler(method string, path string, handler http.Handler) {
	r.Router.Handler(method, path, handler)
}

// HandlerFunc exposes the httprouter HandlerFunc method.
// NOTE: does not handle erors nor panics.
func (r *Router) HandlerFunc(method string, path string, handler http.HandlerFunc) {
	r.Router.HandlerFunc(method, path, handler)
}

// Lookup exposes the httprouter Lookup method.
// NOTE: Return the wrapped handler.
func (r *Router) Lookup(method string, path string) (httprouter.Handle, httprouter.Params, bool) {
	return r.Router.Lookup(method, path)
}

// ServeFiles exposes the httprouter ServeFiles method.
// NOTE: does not handle errors nor panics.
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.Router.ServeFiles(path, root)
}
