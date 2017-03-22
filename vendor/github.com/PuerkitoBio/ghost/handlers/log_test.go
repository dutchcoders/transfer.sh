package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

type testCase struct {
	tok string
	fmt string
	rx  *regexp.Regexp
}

func TestLog(t *testing.T) {
	log.SetFlags(0)
	now := time.Now()

	formats := []testCase{
		testCase{"remote-addr",
			"%s",
			regexp.MustCompile(`^127\.0\.0\.1:\d+\n$`),
		},
		testCase{"date",
			"%s",
			regexp.MustCompile(`^` + fmt.Sprintf("%04d-%02d-%02d", now.Year(), now.Month(), now.Day()) + `\n$`),
		},
		testCase{"method",
			"%s",
			regexp.MustCompile(`^GET\n$`),
		},
		testCase{"url",
			"%s",
			regexp.MustCompile(`^/\n$`),
		},
		testCase{"http-version",
			"%s",
			regexp.MustCompile(`^1\.1\n$`),
		},
		testCase{"status",
			"%d",
			regexp.MustCompile(`^200\n$`),
		},
		testCase{"referer",
			"%s",
			regexp.MustCompile(`^http://www\.test\.com\n$`),
		},
		testCase{"referrer",
			"%s",
			regexp.MustCompile(`^http://www\.test\.com\n$`),
		},
		testCase{"user-agent",
			"%s",
			regexp.MustCompile(`^Go \d+\.\d+ package http\n$`),
		},
		testCase{"bidon",
			"%s",
			regexp.MustCompile(`^\?\n$`),
		},
		testCase{"response-time",
			"%.3f",
			regexp.MustCompile(`^0\.1\d\d\n$`),
		},
		testCase{"req[Accept-Encoding]",
			"%s",
			regexp.MustCompile(`^gzip\n$`),
		},
		testCase{"res[blah]",
			"%s",
			regexp.MustCompile(`^$`),
		},
		testCase{"tiny",
			Ltiny,
			regexp.MustCompile(`^GET / 200  - 0\.1\d\d s\n$`),
		},
		testCase{"short",
			Lshort,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - GET / HTTP/1\.1 200  - 0\.1\d\d s\n$`),
		},
		testCase{"default",
			Ldefault,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - - \[\d{4}-\d{2}-\d{2}\] "GET / HTTP/1\.1" 200  "http://www\.test\.com" "Go \d+\.\d+ package http"\n$`),
		},
		testCase{"res[Content-Type]",
			"%s",
			regexp.MustCompile(`^text/plain\n$`),
		},
	}
	for _, tc := range formats {
		testLogCase(tc, t)
	}
}

func testLogCase(tc testCase, t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	opts := NewLogOptions(log.Printf, tc.fmt, tc.tok)
	opts.DateFormat = "2006-01-02"
	h := LogHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), opts)

	s := httptest.NewServer(h)
	defer s.Close()
	t.Logf("running %s...", tc.tok)
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Referer", "http://www.test.com")
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	ac := buf.String()
	assertTrue(tc.rx.MatchString(ac), fmt.Sprintf("expected log to match '%s', got '%s'", tc.rx.String(), ac), t)
}

func TestForwardedFor(t *testing.T) {
	rx := regexp.MustCompile(`^1\.1\.1\.1:0 - - \[\d{4}-\d{2}-\d{2}\] "GET / HTTP/1\.1" 200  "http://www\.test\.com" "Go \d+\.\d+ package http"\n$`)

	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	opts := NewLogOptions(log.Printf, Ldefault)
	opts.DateFormat = "2006-01-02"

	h := LogHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), opts)

	s := httptest.NewServer(h)
	defer s.Close()
	t.Logf("running ForwardedFor...")
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Referer", "http://www.test.com")
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	ac := buf.String()
	assertTrue(rx.MatchString(ac), fmt.Sprintf("expected log to match '%s', got '%s'", rx.String(), ac), t)
}

func TestImmediate(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetFlags(0)
	log.SetOutput(buf)
	opts := NewLogOptions(nil, Ltiny)
	opts.Immediate = true
	h := LogHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), opts)
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	ac := buf.String()
	// Since it is Immediate logging, status is still 0 and response time is less than 100ms
	rx := regexp.MustCompile(`GET / 0  - 0\.0\d\d s\n`)
	assertTrue(rx.MatchString(ac), fmt.Sprintf("expected log to match '%s', got '%s'", rx.String(), ac), t)
}

func TestCustom(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetFlags(0)
	log.SetOutput(buf)
	opts := NewLogOptions(nil, "%s %s", "method", "custom")
	opts.CustomTokens["custom"] = func(w http.ResponseWriter, r *http.Request) string {
		return "toto"
	}

	h := LogHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), opts)
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	ac := buf.String()
	rx := regexp.MustCompile(`GET toto`)
	assertTrue(rx.MatchString(ac), fmt.Sprintf("expected log to match '%s', got '%s'", rx.String(), ac), t)
}
