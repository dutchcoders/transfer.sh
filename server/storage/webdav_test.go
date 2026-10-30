package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/webdav"
)

const webdavTestContent = "hello world"

// newWebDAVTestStorage starts an in-memory WebDAV server, wrapped by wrap when
// non nil, and returns a WebDAVStorage connected to it holding one test file
func newWebDAVTestStorage(t *testing.T, wrap func(http.Handler) http.Handler) *WebDAVStorage {
	t.Helper()

	var handler http.Handler = &webdav.Handler{
		FileSystem: webdav.NewMemFS(),
		LockSystem: webdav.NewMemLS(),
	}
	if wrap != nil {
		handler = wrap(handler)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	s, err := NewWebDAVStorage(srv.URL, "/transfer", "user", "pass", nil)
	if err != nil {
		t.Fatalf("NewWebDAVStorage: %v", err)
	}
	if err := s.Put(context.Background(), "token", "file.txt", strings.NewReader(webdavTestContent), "text/plain", uint64(len(webdavTestContent))); err != nil {
		t.Fatalf("Put: %v", err)
	}
	return s
}

// readWebDAVTestFile calls Get with rng and returns the body and content length
func readWebDAVTestFile(t *testing.T, s *WebDAVStorage, rng *Range) (string, uint64) {
	t.Helper()

	rc, contentLength, err := s.Get(context.Background(), "token", "file.txt", rng)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer CloseCheck(rc)

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body), contentLength
}

func TestWebDAVStoresReadsAndDeletesFile(t *testing.T) {
	s := newWebDAVTestStorage(t, nil)
	ctx := context.Background()

	size, err := s.Head(ctx, "token", "file.txt")
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if size != uint64(len(webdavTestContent)) {
		t.Fatalf("Head size = %d, want %d", size, len(webdavTestContent))
	}

	body, contentLength := readWebDAVTestFile(t, s, nil)
	if body != webdavTestContent || contentLength != uint64(len(webdavTestContent)) {
		t.Fatalf("Get = %q (%d bytes), want %q", body, contentLength, webdavTestContent)
	}

	if err := s.Delete(ctx, "token", "file.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Head(ctx, "token", "file.txt"); !s.IsNotExist(err) {
		t.Fatalf("Head after Delete: IsNotExist(%v) = false, want true", err)
	}
}

func TestWebDAVGetServesRanges(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		wantBody     string
		wantRange    string
		wantFullFile bool
	}{
		{name: "bounded", header: "bytes=0-4", wantBody: "hello", wantRange: "bytes 0-4/11"},
		{name: "open ended", header: "bytes=6-", wantBody: "world", wantRange: "bytes 6-10/11"},
		{name: "single last byte", header: "bytes=10-10", wantBody: "d", wantRange: "bytes 10-10/11"},
		{name: "start past end serves whole file", header: "bytes=100-", wantBody: webdavTestContent, wantFullFile: true},
	}

	s := newWebDAVTestStorage(t, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := ParseRange(tt.header)
			body, contentLength := readWebDAVTestFile(t, s, rng)

			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if contentLength != uint64(len(tt.wantBody)) {
				t.Errorf("content length = %d, want %d", contentLength, len(tt.wantBody))
			}
			if got := rng.ContentRange(); got != tt.wantRange {
				t.Errorf("Content-Range = %q, want %q", got, tt.wantRange)
			}
		})
	}
}

func TestWebDAVGetServesRangeWhenServerIgnoresRangeHeader(t *testing.T) {
	// Some WebDAV servers answer 200 with the whole file to a Range request;
	// the requested slice must still be returned.
	s := newWebDAVTestStorage(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("Range")
			next.ServeHTTP(w, r)
		})
	})

	rng := ParseRange("bytes=6-")
	body, contentLength := readWebDAVTestFile(t, s, rng)
	if body != "world" || contentLength != 5 {
		t.Fatalf("Get = %q (%d bytes), want %q (5 bytes)", body, contentLength, "world")
	}
	if got := rng.ContentRange(); got != "bytes 6-10/11" {
		t.Fatalf("Content-Range = %q, want %q", got, "bytes 6-10/11")
	}
}

func TestWebDAVIsNotExistOnlyForMissingFiles(t *testing.T) {
	s := newWebDAVTestStorage(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "broken") {
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	ctx := context.Background()

	_, _, err := s.Get(ctx, "token", "missing.txt", nil)
	if err == nil || !s.IsNotExist(err) {
		t.Errorf("missing file: IsNotExist(%v) = false, want true", err)
	}

	_, _, err = s.Get(ctx, "token", "broken.txt", nil)
	if err == nil || s.IsNotExist(err) {
		t.Errorf("server error: IsNotExist(%v) = true, want false", err)
	}

	if s.IsNotExist(nil) {
		t.Error("IsNotExist(nil) = true, want false")
	}
	if s.IsNotExist(errors.New("plain error")) {
		t.Error("IsNotExist(plain error) = true, want false")
	}
}

func TestObjectPathJoinsWithSlashes(t *testing.T) {
	if got := ObjectPath("/base", "token", "file.txt"); got != "/base/token/file.txt" {
		t.Fatalf("ObjectPath = %q, want %q", got, "/base/token/file.txt")
	}
}
