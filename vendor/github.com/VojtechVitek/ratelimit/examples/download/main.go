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

// wget http://localhost:3333 -q --show-progress
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(ratelimit.DownloadSpeed(ratelimit.IP).Rate(1024, time.Second).LimitBy(redis.New(pool), memory.New()))
	r.Get("/", ServeVideo)

	http.ListenAndServe(":3333", r)
}

func ServeVideo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/Users/vojtechvitek/Desktop/govideo.mov")
}
