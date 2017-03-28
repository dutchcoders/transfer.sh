package main

import (
	"net/http"
	"time"

	"github.com/VojtechVitek/ratelimit"
)

// curl -v http://localhost:3333
func main() {
	middleware := ratelimit.Throttle(1)

	http.ListenAndServe(":3333", middleware(http.HandlerFunc(Work)))
}

func Work(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("working hard...\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	time.Sleep(10 * time.Second)
	w.Write([]byte("done"))
}
