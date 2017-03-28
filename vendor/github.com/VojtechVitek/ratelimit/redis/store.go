package redis

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	PrefixKey      = "ratelimit:"
	ErrUnreachable = errors.New("redis is unreachable")
	RetryAfter     = time.Second
)

const skipOnUnhealthy = 1000

type bucketStore struct {
	pool *redis.Pool

	rate          int
	windowSeconds int
	retryAfter    *time.Time
}

// New creates new in-memory token bucket store.
func New(pool *redis.Pool) *bucketStore {
	return &bucketStore{
		pool: pool,
	}
}

func (s *bucketStore) InitRate(rate int, window time.Duration) {
	s.rate = rate
	s.windowSeconds = int(window / time.Second)
	if s.windowSeconds <= 1 {
		s.windowSeconds = 1
	}
}

// Take implements TokenBucketStore interface. It takes token from a bucket
// referenced by a given key, if available.
func (s *bucketStore) Take(key string) (bool, int, time.Time, error) {
	if s.retryAfter != nil {
		if s.retryAfter.After(time.Now()) {
			return false, 0, time.Time{}, ErrUnreachable
		}
		s.retryAfter = nil
	}
	c := s.pool.Get()
	defer c.Close()

	// Number of tokens in the bucket.
	bucketLen, err := redis.Int(c.Do("LLEN", PrefixKey+key))
	if err != nil {
		next := time.Now().Add(time.Second)
		s.retryAfter = &next
		return false, 0, time.Time{}, err
	}

	// Bucket is full.
	if bucketLen >= s.rate {
		return false, 0, time.Time{}, nil
	}

	if bucketLen > 0 {
		// Bucket most probably exists, try to push a new token into it.
		// If RPUSHX returns 0 (ie. key expired between LLEN and RPUSHX), we need
		// to fall-back to RPUSH without returning error.
		c.Send("MULTI")
		c.Send("RPUSHX", PrefixKey+key, "")
		reply, err := redis.Ints(c.Do("EXEC"))
		if err != nil {
			next := time.Now().Add(time.Second)
			s.retryAfter = &next
			return false, 0, time.Time{}, err
		}
		bucketLen = reply[0]
		if bucketLen > 0 {
			return true, s.rate - bucketLen - 1, time.Time{}, nil
		}
	}

	c.Send("MULTI")
	c.Send("RPUSH", PrefixKey+key, "")
	c.Send("EXPIRE", PrefixKey+key, s.windowSeconds)
	if _, err := c.Do("EXEC"); err != nil {
		next := time.Now().Add(time.Second)
		s.retryAfter = &next
		return false, 0, time.Time{}, err
	}

	return true, s.rate - bucketLen - 1, time.Time{}, nil
}
