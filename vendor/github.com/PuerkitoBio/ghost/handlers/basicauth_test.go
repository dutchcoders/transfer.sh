package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnauth(t *testing.T) {
	h := BasicAuthHandler(StaticFileHandler("./testdata/script.js"), func(u, pwd string) (interface{}, bool) {
		if u == "me" && pwd == "you" {
			return u, true
		}
		return nil, false
	}, "foo")
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusUnauthorized, res.StatusCode, t)
	assertHeader("Www-Authenticate", `Basic realm="foo"`, res, t)
}

func TestGzippedAuth(t *testing.T) {
	h := GZIPHandler(BasicAuthHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			usr, ok := GetUser(w)
			if assertTrue(ok, "expected authenticated user, got false", t) {
				assertTrue(usr.(string) == "meyou", fmt.Sprintf("expected user data to be 'meyou', got '%s'", usr), t)
			}
			usr, ok = GetUserName(w)
			if assertTrue(ok, "expected authenticated user name, got false", t) {
				assertTrue(usr == "me", fmt.Sprintf("expected user name to be 'me', got '%s'", usr), t)
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(usr.(string)))
		}), func(u, pwd string) (interface{}, bool) {
		if u == "me" && pwd == "you" {
			return u + pwd, true
		}
		return nil, false
	}, ""), nil)

	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", "http://me:you@"+s.URL[7:], nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertGzippedBody([]byte("me"), res, t)
}
