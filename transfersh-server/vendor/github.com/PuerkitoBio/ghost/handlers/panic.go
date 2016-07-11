package handlers

import (
	"fmt"
	"net/http"
)

// Augmented response writer to hold the panic data (can be anything, not necessarily an error
// interface).
type errResponseWriter struct {
	http.ResponseWriter
	perr interface{}
}

// Implement the WrapWriter interface.
func (this *errResponseWriter) WrappedWriter() http.ResponseWriter {
	return this.ResponseWriter
}

// PanicHandlerFunc is the same as PanicHandler, it is just a convenience
// signature that accepts a func(http.ResponseWriter, *http.Request) instead of
// a http.Handler interface. It saves the boilerplate http.HandlerFunc() cast.
func PanicHandlerFunc(h http.HandlerFunc, errH http.HandlerFunc) http.HandlerFunc {
	return PanicHandler(h, errH)
}

// Calls the wrapped handler and on panic calls the specified error handler. If the error handler is nil,
// responds with a 500 error message.
func PanicHandler(h http.Handler, errH http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if errH != nil {
					ew := &errResponseWriter{w, err}
					errH.ServeHTTP(ew, r)
				} else {
					http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
				}
			}
		}()

		// Call the protected handler
		h.ServeHTTP(w, r)
	}
}

// Helper function to retrieve the panic error, if any.
func GetPanicError(w http.ResponseWriter) (interface{}, bool) {
	er, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*errResponseWriter)
		return ok
	})
	if ok {
		return er.(*errResponseWriter).perr, true
	}
	return nil, false
}
