package ratelimit_test

import (
	"net/http"
	"time"

	"github.com/VojtechVitek/ratelimit"
)

func ExampleThrottle() {
	middleware := ratelimit.Throttle(1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("working hard...\n\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		time.Sleep(10 * time.Second)
		w.Write([]byte("done"))
	})

	http.ListenAndServe(":3333", middleware(handler))
}
