package main

import (
	"net/http"
	"time"

	"github.com/VojtechVitek/ratelimit"
	"github.com/VojtechVitek/ratelimit/memory"
	"github.com/VojtechVitek/ratelimit/redis"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

var pool = &redigo.Pool{
	MaxIdle:     10,
	MaxActive:   50,
	IdleTimeout: 300 * time.Second,
	Wait:        false, // Important
	Dial: func() (redigo.Conn, error) {
		c, err := redigo.DialTimeout("tcp", "127.0.0.1:6379", 200*time.Millisecond, 100*time.Millisecond, 100*time.Millisecond)
		if err != nil {
			return nil, err
		}
		return c, err
	},
	TestOnBorrow: func(c redigo.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

// while :; do curl -v localhost:3333; sleep 0.1; done
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//r.Use(ratelimit.Request(ratelimit.IP).Rate(1, time.Second).LimitBy(redis.New(pool)))

	r.Use(ratelimit.Request(ratelimit.IP).Rate(5, 5*time.Second).LimitBy(redis.New(pool), memory.New()))

	r.Get("/", Hello)

	http.ListenAndServe(":3333", r)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!\n"))
	//w.Write([]byte("Hello user_id=" + r.URL.Query().Get("user_id") + "\n"))
}

func UserKey(r *http.Request) string {
	user := r.URL.Query().Get("user_id")
	// user, _ := r.Context().Value("session.user_id").(string)
	return user
}
