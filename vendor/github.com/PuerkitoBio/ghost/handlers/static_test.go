package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeFile(t *testing.T) {
	h := StaticFileHandler("./testdata/styles.css")
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Type", "text/css; charset=utf-8", res, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(`* {
  background-color: white;
}`), res, t)
}

func TestGzippedFile(t *testing.T) {
	h := GZIPHandler(StaticFileHandler("./testdata/styles.css"), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "*")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "gzip", res, t)
	assertHeader("Content-Type", "text/css; charset=utf-8", res, t)
	assertGzippedBody([]byte(`* {
  background-color: white;
}`), res, t)
}
