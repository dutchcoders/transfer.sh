package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestFavicon(t *testing.T) {
	s := httptest.NewServer(FaviconHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}, "./testdata/favicon.ico", time.Second))
	defer s.Close()

	res, err := http.Get(s.URL + "/favicon.ico")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Type", "image/x-icon", res, t)
	assertHeader("Cache-Control", "public, max-age=1", res, t)
	assertHeader("Content-Length", "1406", res, t)
}

func TestFaviconInvalidPath(t *testing.T) {
	s := httptest.NewServer(FaviconHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}, "./testdata/xfavicon.ico", time.Second))
	defer s.Close()

	res, err := http.Get(s.URL + "/favicon.ico")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	assertStatus(http.StatusNotFound, res.StatusCode, t)
}

func TestFaviconFromCache(t *testing.T) {
	s := httptest.NewServer(FaviconHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}, "./testdata/favicon.ico", time.Second))
	defer s.Close()

	res, err := http.Get(s.URL + "/favicon.ico")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Rename the file temporarily
	err = os.Rename("./testdata/favicon.ico", "./testdata/xfavicon.ico")
	if err != nil {
		panic(err)
	}
	defer os.Rename("./testdata/xfavicon.ico", "./testdata/favicon.ico")
	res, err = http.Get(s.URL + "/favicon.ico")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Type", "image/x-icon", res, t)
	assertHeader("Cache-Control", "public, max-age=1", res, t)
	assertHeader("Content-Length", "1406", res, t)
}
