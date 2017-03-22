package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanic(t *testing.T) {
	h := PanicHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("test")
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusInternalServerError, res.StatusCode, t)
}

func TestNoPanic(t *testing.T) {
	h := PanicHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
}

func TestPanicCustom(t *testing.T) {
	h := PanicHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("ok")
		}),
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				err, ok := GetPanicError(w)
				if !ok {
					panic("no panic error found")
				}
				w.WriteHeader(501)
				w.Write([]byte(err.(string)))
			}))
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(501, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
}
