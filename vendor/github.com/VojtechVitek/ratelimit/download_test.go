package ratelimit_test

import (
	"net/http"
	"time"

	"github.com/VojtechVitek/ratelimit"
	"github.com/VojtechVitek/ratelimit/memory"
)

// Watch the download speed with
// wget http://localhost:3333/file -q --show-progress
func ExampleDownloadSpeed() {
	middleware := ratelimit.DownloadSpeed(ratelimit.IP).Rate(1024, time.Second).LimitBy(memory.New())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/dev/random")
	})

	http.ListenAndServe(":3333", middleware(handler))
}
