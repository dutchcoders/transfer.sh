package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipped(t *testing.T) {
	body := "This is the body"
	headers := []string{"gzip", "*", "gzip, deflate, sdch"}

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	for _, hdr := range headers {
		t.Logf("running with Accept-Encoding header %s", hdr)
		req, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Accept-Encoding", hdr)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		assertStatus(http.StatusOK, res.StatusCode, t)
		assertHeader("Content-Encoding", "gzip", res, t)
		assertGzippedBody([]byte(body), res, t)
	}
}

func TestNoGzip(t *testing.T) {
	body := "This is the body"

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestGzipOuterPanic(t *testing.T) {
	msg := "ko"

	h := PanicHandler(
		GZIPHandler(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				panic(msg)
			}), nil), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusInternalServerError, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(msg+"\n"), res, t)
}

func TestNoGzipOnFilter(t *testing.T) {
	body := "This is the body"

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "x/x")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestNoGzipOnCustomFilter(t *testing.T) {
	body := "This is the body"

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), func(w http.ResponseWriter, r *http.Request) bool {
		return false
	})
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestGzipOnCustomFilter(t *testing.T) {
	body := "This is the body"

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "x/x")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), func(w http.ResponseWriter, r *http.Request) bool {
		return true
	})
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "gzip", res, t)
	assertGzippedBody([]byte(body), res, t)
}
