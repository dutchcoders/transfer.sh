package handlers

import (
	"net/http"
)

// Structure that holds the context map and exposes the ResponseWriter interface.
type contextResponseWriter struct {
	http.ResponseWriter
	m map[interface{}]interface{}
}

// Implement the WrapWriter interface.
func (this *contextResponseWriter) WrappedWriter() http.ResponseWriter {
	return this.ResponseWriter
}

// ContextHandlerFunc is the same as ContextHandler, it is just a convenience
// signature that accepts a func(http.ResponseWriter, *http.Request) instead of
// a http.Handler interface. It saves the boilerplate http.HandlerFunc() cast.
func ContextHandlerFunc(h http.HandlerFunc, cap int) http.HandlerFunc {
	return ContextHandler(h, cap)
}

// ContextHandler gives a context storage that lives only for the duration of
// the request, with no locking involved.
func ContextHandler(h http.Handler, cap int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := GetContext(w); ok {
			// Self-awareness, context handler is already set up
			h.ServeHTTP(w, r)
			return
		}

		// Create the context-providing ResponseWriter replacement.
		ctxw := &contextResponseWriter{
			w,
			make(map[interface{}]interface{}, cap),
		}
		// Call the wrapped handler with the context-aware writer
		h.ServeHTTP(ctxw, r)
	}
}

// Helper function to retrieve the context map from the ResponseWriter interface.
func GetContext(w http.ResponseWriter) (map[interface{}]interface{}, bool) {
	ctxw, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*contextResponseWriter)
		return ok
	})
	if ok {
		return ctxw.(*contextResponseWriter).m, true
	}
	return nil, false
}
