// Copyright 2015 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package talksapp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/memcache"

	"github.com/golang/gddo/gosrc"
)

const importPath = "github.com/user/repo/path/to/presentation.slide"

func TestHome(t *testing.T) {
	do(t, "GET", "/", func(r *http.Request) {
		w := httptest.NewRecorder()
		handlerFunc(serveRoot).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status: %d, got: %d", http.StatusOK, w.Code)
		}

		if !strings.Contains(w.Body.String(), "go-talks.appspot.org") {
			t.Fatal("expected response to contain: go-talks.appspot.org")
		}
	})
}

func TestPresentation(t *testing.T) {
	presentationTitle := "My awesome presentation!"
	presentationSrc := []byte(presentationTitle + `

Subtitle

* Slide 1

- Foo
- Bar
- Baz
`)

	originalGetPresentation := getPresentation
	getPresentation = func(client *http.Client, importPath string) (*gosrc.Presentation, error) {
		return &gosrc.Presentation{
			Filename: "presentation.slide",
			Files: map[string][]byte{
				"presentation.slide": presentationSrc,
			},
		}, nil
	}
	defer func() {
		getPresentation = originalGetPresentation
	}()

	do(t, "GET", "/"+importPath, func(r *http.Request) {
		w := httptest.NewRecorder()
		handlerFunc(serveRoot).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status: %d, got: %d", http.StatusOK, w.Code)
		}

		if !strings.Contains(w.Body.String(), presentationTitle) {
			t.Fatalf("unexpected response body: %s", w.Body)
		}

		c := appengine.NewContext(r)
		_, err := memcache.Get(c, importPath)

		if err == memcache.ErrCacheMiss {
			t.Fatal("expected result to be cached")
		}

		if err != nil {
			t.Fatalf("expected no error, got: %s", err)
		}
	})
}

func TestPresentationCacheHit(t *testing.T) {
	do(t, "GET", "/"+importPath, func(r *http.Request) {
		cachedPresentation := "<div>My Presentation</div>"

		c := appengine.NewContext(r)
		memcache.Add(c, &memcache.Item{
			Key:        importPath,
			Value:      []byte(cachedPresentation),
			Expiration: time.Hour,
		})

		w := httptest.NewRecorder()
		handlerFunc(serveRoot).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status: %d, got: %d", http.StatusOK, w.Code)
		}

		if w.Body.String() != cachedPresentation {
			t.Fatal("response does not matched cached presentation")
		}
	})
}

func TestPresentationNotFound(t *testing.T) {
	originalGetPresentation := getPresentation
	getPresentation = func(client *http.Client, importPath string) (*gosrc.Presentation, error) {
		return nil, gosrc.NotFoundError{}
	}
	defer func() {
		getPresentation = originalGetPresentation
	}()

	do(t, "GET", "/"+importPath, func(r *http.Request) {
		w := httptest.NewRecorder()
		handlerFunc(serveRoot).ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status: %d, got: %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestWrongMethod(t *testing.T) {
	do(t, "POST", "/", func(r *http.Request) {
		w := httptest.NewRecorder()
		handlerFunc(serveRoot).ServeHTTP(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d", http.StatusMethodNotAllowed)
		}
	})
}

func TestCompile(t *testing.T) {
	version := "2"
	body := `
	package main

	import "fmt"

	func main() {
		fmt.fmtPrintln("Hello, playground")
	}
	`
	responseJSON := `{"Errors":"","Events":[{"Message":"Hello, playground\n","Kind":"stdout","Delay":0}]}`

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			formVersion := r.FormValue("version")
			formBody := r.FormValue("body")

			if formVersion != version {
				t.Fatalf("expected version sent to play.golang.org to be: %s, was: %s", version, formVersion)
			}

			if formBody != body {
				t.Fatalf("expected body sent to play.golang.org to be: %s, was: %s", body, formBody)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)

			fmt.Fprintln(w, responseJSON)
		}),
	)
	defer server.Close()

	defer func(old string) { playCompileURL = old }(playCompileURL)
	playCompileURL = server.URL

	do(t, "POST", "/compile", func(r *http.Request) {
		r.PostForm = url.Values{
			"version": []string{version},
			"body":    []string{body},
		}

		w := httptest.NewRecorder()
		handlerFunc(serveCompile).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status: %d, got: %d", http.StatusOK, w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if w.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected Content-Type: %s", contentType)
		}

		if strings.TrimSpace(w.Body.String()) != responseJSON {
			t.Fatalf("unexpected response body: %s", w.Body)
		}
	})
}

func TestBot(t *testing.T) {
	do(t, "GET", "/bot.html", func(r *http.Request) {
		w := httptest.NewRecorder()
		handlerFunc(serveBot).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status: %d, got: %d", http.StatusOK, w.Code)
		}

		if !strings.Contains(w.Body.String(), contactEmail) {
			t.Fatalf("expected body to contain %s", contactEmail)
		}
	})
}

func do(t *testing.T, method, path string, f func(*http.Request)) {
	i, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer i.Close()

	r, err := i.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal(err)
	}

	f(r)
}
