package memory

import (
	"sync"
	"time"
)

type token struct{}

type bucketStore struct {
	sync.Mutex // guards buckets
	buckets    map[string]chan token
	bucketLen  int
	reset      time.Time
}

// New creates new in-memory token bucket store.
func New() *bucketStore {
	return &bucketStore{
		buckets: map[string]chan token{},
	}
}

func (s *bucketStore) InitRate(rate int, window time.Duration) {
	s.bucketLen = rate
	s.reset = time.Now()

	go func() {
		interval := time.Duration(int(window) / rate)
		tick := time.NewTicker(interval)
		for t := range tick.C {
			s.Lock()
			s.reset = t.Add(interval)
			for key, bucket := range s.buckets {
				select {
				case <-bucket:
				default:
					delete(s.buckets, key)
				}
			}
			s.Unlock()
		}
	}()
}

// Take implements TokenBucketStore interface. It takes token from a bucket
// referenced by a given key, if available.
func (s *bucketStore) Take(key string) (bool, int, time.Time, error) {
	s.Lock()
	bucket, ok := s.buckets[key]
	if !ok {
		bucket = make(chan token, s.bucketLen)
		s.buckets[key] = bucket
	}
	s.Unlock()
	select {
	case bucket <- token{}:
		return true, cap(bucket) - len(bucket), s.reset, nil
	default:
		return false, 0, s.reset, nil
	}
}
