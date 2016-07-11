package handlers

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	ErrNoKeyPrefix = errors.New("cannot get session keys without a key prefix")
)

type RedisStoreOptions struct {
	Network              string
	Address              string
	ConnectTimeout       time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	Database             int           // Redis database to use for session keys
	KeyPrefix            string        // If set, keys will be KeyPrefix:SessionID (semicolon added)
	BrowserSessServerTTL time.Duration // Defaults to 2 days
}

type RedisStore struct {
	opts *RedisStoreOptions
	conn redis.Conn
}

// Create a redis session store with the specified options.
func NewRedisStore(opts *RedisStoreOptions) *RedisStore {
	var err error
	rs := &RedisStore{opts, nil}
	rs.conn, err = redis.DialTimeout(opts.Network, opts.Address, opts.ConnectTimeout,
		opts.ReadTimeout, opts.WriteTimeout)
	if err != nil {
		panic(err)
	}
	return rs
}

// Get the session from the store.
func (this *RedisStore) Get(id string) (*Session, error) {
	key := id
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + id
	}
	b, err := redis.Bytes(this.conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	var sess Session
	err = json.Unmarshal(b, &sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// Save the session into the store.
func (this *RedisStore) Set(sess *Session) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	key := sess.ID()
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + sess.ID()
	}
	ttl := sess.MaxAge()
	if ttl == 0 {
		// Browser session, set to specified TTL
		ttl = this.opts.BrowserSessServerTTL
		if ttl == 0 {
			ttl = 2 * 24 * time.Hour // Default to 2 days
		}
	}
	_, err = this.conn.Do("SETEX", key, int(ttl.Seconds()), b)
	if err != nil {
		return err
	}
	return nil
}

// Delete the session from the store.
func (this *RedisStore) Delete(id string) error {
	key := id
	if this.opts.KeyPrefix != "" {
		key = this.opts.KeyPrefix + ":" + id
	}
	_, err := this.conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

// Clear all sessions from the store. Requires the use of a key
// prefix in the store options, otherwise the method refuses to delete all keys.
func (this *RedisStore) Clear() error {
	vals, err := this.getSessionKeys()
	if err != nil {
		return err
	}
	if len(vals) > 0 {
		this.conn.Send("MULTI")
		for _, v := range vals {
			this.conn.Send("DEL", v)
		}
		_, err = this.conn.Do("EXEC")
		if err != nil {
			return err
		}
	}
	return nil
}

// Get the number of session keys in the store. Requires the use of a
// key prefix in the store options, otherwise returns -1 (cannot tell
// session keys from other keys).
func (this *RedisStore) Len() int {
	vals, err := this.getSessionKeys()
	if err != nil {
		return -1
	}
	return len(vals)
}

func (this *RedisStore) getSessionKeys() ([]interface{}, error) {
	if this.opts.KeyPrefix != "" {
		return redis.Values(this.conn.Do("KEYS", this.opts.KeyPrefix+":*"))
	}
	return nil, ErrNoKeyPrefix
}
