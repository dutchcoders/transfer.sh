package ratelimit

import (
	"net/http"
	"time"
)

// TokenBucketStore is an interface for for any storage implementing
// Token Bucket algorithm.
type TokenBucketStore interface {
	InitRate(rate int, window time.Duration)
	Take(key string) (taken bool, remaining int, reset time.Time, err error)
}

// KeyFn is a function returning bucket key depending on request data.
type KeyFn func(r *http.Request) string
