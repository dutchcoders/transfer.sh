package ratelimit

import (
	"net/http"
	"time"
)

func DownloadSpeed(keyFn KeyFn) *downloadBuilder {
	return &downloadBuilder{
		keyFn: keyFn,
	}
}

type downloadBuilder struct {
	keyFn  KeyFn
	rate   int
	window time.Duration
}

func (b *downloadBuilder) Rate(rate int, window time.Duration) *downloadBuilder {
	b.rate = rate
	b.window = window
	return b
}

// TODO: Custom burst?
// func (b *downloadBuilder) Burst(burst int) *downloadBuilder {}

func (b *downloadBuilder) LimitBy(store TokenBucketStore, fallbackStores ...TokenBucketStore) func(http.Handler) http.Handler {
	store.InitRate(b.rate, b.window)
	for _, store := range fallbackStores {
		store.InitRate(b.rate, b.window)
	}

	downloadLimiter := downloadLimiter{
		downloadBuilder: b,
		store:           store,
		fallbackStores:  fallbackStores,
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			key := downloadLimiter.keyFn(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			lw := &limitWriter{
				ResponseWriter:  w,
				downloadLimiter: &downloadLimiter,
				key:             key,
			}

			next.ServeHTTP(lw, r)
		}
		return http.HandlerFunc(fn)
	}
}

type downloadLimiter struct {
	*downloadBuilder

	next           http.Handler
	store          TokenBucketStore
	fallbackStores []TokenBucketStore
}

type limitWriter struct {
	http.ResponseWriter
	*downloadLimiter

	key         string
	wroteHeader bool
	canWrite    int64
}

func (w *limitWriter) Write(buf []byte) (int, error) {
	total := 0
	for {
		if w.canWrite < 1024 {
			ok, _, _, err := w.downloadLimiter.store.Take("download:" + w.key)
			if err != nil {
				for _, store := range w.fallbackStores {
					ok, _, _, err = store.Take("download:" + w.key)
					if err == nil {
						break
					}
				}
			}
			if err != nil {
				return total, err
			}
			if ok {
				w.canWrite += 1024
			}
		}
		if w.canWrite == 0 {
			continue
		}

		max := len(buf) - total
		if int(w.canWrite) < max {
			max = int(w.canWrite)
		}
		if max == 0 {
			return total, nil
		}

		n, err := w.ResponseWriter.Write(buf[total : total+max])
		w.canWrite -= int64(n)
		total += n
		if err != nil {
			return total, err
		}
	}
}
