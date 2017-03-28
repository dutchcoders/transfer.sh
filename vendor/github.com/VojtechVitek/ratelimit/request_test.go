package ratelimit_test

import (
	"net/http"
	"time"

	"github.com/VojtechVitek/ratelimit"
	"github.com/VojtechVitek/ratelimit/memory"
)

func ExampleRequest() {
	middleware := ratelimit.Request(ratelimit.IP).Rate(30, time.Minute).LimitBy(memory.New())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	http.ListenAndServe(":3333", middleware(handler))
}
