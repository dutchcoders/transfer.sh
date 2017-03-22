package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	key := "key"
	val := 10
	body := "this is the output"

	h2 := wrappedHandler(t, key, val, body)
	// Create the context handler with a wrapped handler
	h := ContextHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, _ := GetContext(w)
			assertTrue(ctx != nil, "expected context to be non-nil", t)
			assertTrue(len(ctx) == 0, fmt.Sprintf("expected context to be empty, got %d", len(ctx)), t)
			ctx[key] = val
			h2.ServeHTTP(w, r)
		}), 2)
	s := httptest.NewServer(h)
	defer s.Close()

	// First call
	res, err := http.DefaultClient.Get(s.URL)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	// Second call, context should be cleaned at start
	res, err = http.DefaultClient.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte(body), res, t)
}

func TestWrappedContext(t *testing.T) {
	key := "key"
	val := 10
	body := "this is the output"

	h2 := wrappedHandler(t, key, val, body)
	h := ContextHandler(LogHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, _ := GetContext(w)
			if !assertTrue(ctx != nil, "expected context to be non-nil", t) {
				panic("ctx is nil")
			}
			assertTrue(len(ctx) == 0, fmt.Sprintf("expected context to be empty, got %d", len(ctx)), t)
			ctx[key] = val
			h2.ServeHTTP(w, r)
		}), NewLogOptions(nil, "%s", "url")), 2)
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.DefaultClient.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte(body), res, t)
}

func wrappedHandler(t *testing.T, k, v interface{}, body string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, _ := GetContext(w)
			ac := ctx[k]
			assertTrue(ac == v, fmt.Sprintf("expected value to be %v, got %v", v, ac), t)

			// Actually write something
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		})
}
