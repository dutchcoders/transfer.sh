package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var (
	_ = Suite(&SuiteRedirectWithForceHTTPs{})
	_ = Suite(&SuiteRedirectWithoutForceHTTPs{})
)

type SuiteRedirectWithForceHTTPs struct {
	handler http.HandlerFunc
}

func (s *SuiteRedirectWithForceHTTPs) SetUpTest(c *C) {
	srvr, err := New(ForceHTTPs())
	c.Assert(err, IsNil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})

	s.handler = srvr.RedirectHandler(handler)
}

func (s *SuiteRedirectWithForceHTTPs) TestHTTPs(c *C) {
	req := httptest.NewRequest("GET", "https://test/test", nil)

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
}

func (s *SuiteRedirectWithForceHTTPs) TestOnion(c *C) {
	req := httptest.NewRequest("GET", "http://test.onion/test", nil)

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
}

func (s *SuiteRedirectWithForceHTTPs) TestXForwardedFor(c *C) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
}

func (s *SuiteRedirectWithForceHTTPs) TestHTTP(c *C) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/test", nil)

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusPermanentRedirect)
	c.Assert(resp.Header.Get("Location"), Equals, "https://127.0.0.1/test")
}

type SuiteRedirectWithoutForceHTTPs struct {
	handler http.HandlerFunc
}

func (s *SuiteRedirectWithoutForceHTTPs) SetUpTest(c *C) {
	srvr, err := New()
	c.Assert(err, IsNil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})

	s.handler = srvr.RedirectHandler(handler)
}

func (s *SuiteRedirectWithoutForceHTTPs) TestHTTP(c *C) {
	req := httptest.NewRequest("GET", "http://127.0.0.1/test", nil)

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
}

func (s *SuiteRedirectWithoutForceHTTPs) TestHTTPs(c *C) {
	req := httptest.NewRequest("GET", "https://127.0.0.1/test", nil)

	w := httptest.NewRecorder()
	s.handler(w, req)

	resp := w.Result()
	c.Assert(resp.StatusCode, Equals, http.StatusOK)
}
