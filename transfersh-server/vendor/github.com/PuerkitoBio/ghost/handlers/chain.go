package handlers

import (
	"net/http"
)

// ChainableHandler is a valid Handler interface, and adds the possibility to
// chain other handlers.
type ChainableHandler interface {
	http.Handler
	Chain(http.Handler) ChainableHandler
	ChainFunc(http.HandlerFunc) ChainableHandler
}

// Default implementation of a simple ChainableHandler
type chainHandler struct {
	http.Handler
}

func (this *chainHandler) ChainFunc(h http.HandlerFunc) ChainableHandler {
	return this.Chain(h)
}

// Implementation of the ChainableHandler interface, calls the chained handler
// after the current one (sequential).
func (this *chainHandler) Chain(h http.Handler) ChainableHandler {
	return &chainHandler{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add the chained handler after the call to this handler
			this.ServeHTTP(w, r)
			h.ServeHTTP(w, r)
		}),
	}
}

// Convert a standard http handler to a chainable handler interface.
func NewChainableHandler(h http.Handler) ChainableHandler {
	return &chainHandler{
		h,
	}
}

// Helper function to chain multiple handler functions in a single call.
func ChainHandlerFuncs(h ...http.HandlerFunc) ChainableHandler {
	return &chainHandler{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, v := range h {
				v(w, r)
			}
		}),
	}
}

// Helper function to chain multiple handlers in a single call.
func ChainHandlers(h ...http.Handler) ChainableHandler {
	return &chainHandler{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, v := range h {
				v.ServeHTTP(w, r)
			}
		}),
	}
}
