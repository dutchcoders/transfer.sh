package ratelimit

import "net/http"

// Throttle is a middleware that limits number of currently
// processed requests at a time.
func Throttle(limit int) func(http.Handler) http.Handler {
	if limit <= 0 {
		panic("Throttle expects limit > 0")
	}

	t := throttler{
		tokens: make(chan token, limit),
	}
	for i := 0; i < limit; i++ {
		t.tokens <- token{}
	}

	fn := func(h http.Handler) http.Handler {
		t.h = h
		return &t
	}

	return fn
}

// token represents a request that is being processed.
type token struct{}

// throttler limits number of currently processed requests at a time.
type throttler struct {
	h      http.Handler
	tokens chan token
}

// ServeHTTP implements http.Handler interface.
func (t *throttler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		return
	case tok := <-t.tokens:
		defer func() {
			t.tokens <- tok
		}()
		t.h.ServeHTTP(w, r)
	}
}
