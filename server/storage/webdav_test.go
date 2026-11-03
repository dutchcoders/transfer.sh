package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/webdav"
)

const webdavTestContent = "hello world"

// newWebDAVTestStorage starts an in-memory WebDAV server, wrapped by wrap when
// non nil, and returns a WebDAVStorage connected to it holding one test file
func newWebDAVTestStorage(t *testing.T, wrap func(http.Handler) http.Handler) *WebDAVStorage {
	t.Helper()

	s := serveWebDAVTestStorage(t, webdav.NewMemFS(), wrap)
	if err := s.Put(context.Background(), "token", "file.txt", strings.NewReader(webdavTestContent), "text/plain", uint64(len(webdavTestContent))); err != nil {
		t.Fatalf("Put: %v", err)
	}
	return s
}

// serveWebDAVTestStorage serves fs over WebDAV, wrapped by wrap when non nil,
// and returns a WebDAVStorage rooted at /transfer on that server
func serveWebDAVTestStorage(t *testing.T, fs webdav.FileSystem, wrap func(http.Handler) http.Handler) *WebDAVStorage {
	t.Helper()

	var handler http.Handler = &webdav.Handler{
		FileSystem: fs,
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

// statusRecorder stores the status code a handler answers with. It is
// stored before any byte of the body goes out, and atomically, since the test
// reads it from another goroutine once the client has the response.
type statusRecorder struct {
	http.ResponseWriter
	status *atomic.Int32
}

// WriteHeader records code before passing it on
func (r statusRecorder) WriteHeader(code int) {
	r.status.CompareAndSwap(0, int32(code))
	r.ResponseWriter.WriteHeader(code)
}

// Write records the implicit 200 of a handler that never calls WriteHeader
func (r statusRecorder) Write(b []byte) (int, error) {
	r.status.CompareAndSwap(0, http.StatusOK)
	return r.ResponseWriter.Write(b)
}

func TestWebDAVGetServesRangeWhenServerIgnoresRangeHeader(t *testing.T) {
	// Some WebDAV servers answer 200 with the whole file to a Range request
	// (mod_dav with gzip on, proxies that strip Range); the requested slice
	// must still be returned. x/net/webdav does honor Range and would answer
	// 206, so the wrapper below deletes the header before the server sees it.
	// That is what makes this a 200 scenario, hence the assertion on the
	// status: without it, a change in the wrapper could silently turn this
	// back into a plain 206 test.
	var getStatus atomic.Int32
	s := newWebDAVTestStorage(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("Range")
			if r.Method == http.MethodGet {
				w = statusRecorder{ResponseWriter: w, status: &getStatus}
			}
			next.ServeHTTP(w, r)
		})
	})

	rng := ParseRange("bytes=6-")
	body, contentLength := readWebDAVTestFile(t, s, rng)
	t.Logf("WebDAV server answered the Range GET with %d", getStatus.Load())
	if getStatus := getStatus.Load(); getStatus != http.StatusOK {
		t.Fatalf("WebDAV server status = %d, want %d: the Range header reached it", getStatus, http.StatusOK)
	}
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

// newWebDAVPurgeStorage serves a temporary directory over WebDAV, wrapped by
// wrap when non nil. The in-memory file system cannot backdate files, a
// directory on disk can, through putAgedWebDAVTestFile.
func newWebDAVPurgeStorage(t *testing.T, wrap func(http.Handler) http.Handler) (*WebDAVStorage, string) {
	t.Helper()

	root := t.TempDir()
	return serveWebDAVTestStorage(t, webdav.Dir(root), wrap), root
}

// putAgedWebDAVTestFile uploads token/filename and sets its modification
// time to age ago on the disk behind the WebDAV server
func putAgedWebDAVTestFile(t *testing.T, s *WebDAVStorage, root, token, filename string, age time.Duration) {
	t.Helper()

	if err := s.Put(context.Background(), token, filename, strings.NewReader(webdavTestContent), "text/plain", uint64(len(webdavTestContent))); err != nil {
		t.Fatalf("Put %s/%s: %v", token, filename, err)
	}
	mtime := time.Now().Add(-age)
	if err := os.Chtimes(filepath.Join(root, "transfer", token, filename), mtime, mtime); err != nil {
		t.Fatalf("Chtimes %s/%s: %v", token, filename, err)
	}
}

// webdavTestFileExists reports whether token/filename can still be found
func webdavTestFileExists(t *testing.T, s *WebDAVStorage, token, filename string) bool {
	t.Helper()

	_, err := s.Head(context.Background(), token, filename)
	if err != nil && !s.IsNotExist(err) {
		t.Fatalf("Head %s/%s: %v", token, filename, err)
	}
	return err == nil
}

func TestWebDAVPurgeDeletesOnlyExpiredFiles(t *testing.T) {
	s, root := newWebDAVPurgeStorage(t, nil)
	putAgedWebDAVTestFile(t, s, root, "old", "file.txt", 10*24*time.Hour)
	putAgedWebDAVTestFile(t, s, root, "old", "file.txt.metadata", 10*24*time.Hour)
	putAgedWebDAVTestFile(t, s, root, "new", "file.txt", time.Hour)
	putAgedWebDAVTestFile(t, s, root, "new", "file.txt.metadata", time.Hour)

	if err := s.Purge(context.Background(), 7*24*time.Hour); err != nil {
		t.Fatalf("Purge: %v", err)
	}

	for _, name := range []string{"file.txt", "file.txt.metadata"} {
		if webdavTestFileExists(t, s, "old", name) {
			t.Errorf("old/%s still exists after purge", name)
		}
		if !webdavTestFileExists(t, s, "new", name) {
			t.Errorf("new/%s was purged, it is only an hour old", name)
		}
	}
}

func TestWebDAVPurgeWithoutBaseDirectoryIsNoop(t *testing.T) {
	// Nothing uploaded yet: the base directory does not exist on the server.
	s, _ := newWebDAVPurgeStorage(t, nil)

	if err := s.Purge(context.Background(), 24*time.Hour); err != nil {
		t.Fatalf("Purge on a missing base directory: %v", err)
	}
}

func TestWebDAVPurgeKeepsFilesWithoutModificationTime(t *testing.T) {
	// Strip getlastmodified from every PROPFIND answer, as some servers do.
	lastModified := regexp.MustCompile(`<D:getlastmodified>[^<]*</D:getlastmodified>`)
	s, root := newWebDAVPurgeStorage(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PROPFIND" {
				next.ServeHTTP(w, r)
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			for k, v := range rec.Header() {
				if k != "Content-Length" {
					w.Header()[k] = v
				}
			}
			w.WriteHeader(rec.Code)
			_, _ = w.Write(lastModified.ReplaceAll(rec.Body.Bytes(), nil))
		})
	})
	putAgedWebDAVTestFile(t, s, root, "token", "file.txt", 10*24*time.Hour)

	if err := s.Purge(context.Background(), 24*time.Hour); err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if !webdavTestFileExists(t, s, "token", "file.txt") {
		t.Error("file without a modification time was purged")
	}
}

func TestWebDAVPurgeReportsServerErrors(t *testing.T) {
	s, root := newWebDAVPurgeStorage(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PROPFIND" && strings.Contains(r.URL.Path, "token") {
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	putAgedWebDAVTestFile(t, s, root, "token", "file.txt", 10*24*time.Hour)

	err := s.Purge(context.Background(), 24*time.Hour)
	if err == nil || s.IsNotExist(err) {
		t.Fatalf("Purge = %v, want a server error", err)
	}
}

func TestWebDAVPurgeStopsWhenContextIsCancelled(t *testing.T) {
	s, root := newWebDAVPurgeStorage(t, nil)
	putAgedWebDAVTestFile(t, s, root, "token", "file.txt", 10*24*time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Purge(ctx, 24*time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("Purge = %v, want context.Canceled", err)
	}
	if !webdavTestFileExists(t, s, "token", "file.txt") {
		t.Error("file was purged despite the cancelled context")
	}
}

// writeAgedDiskFile creates rel under the WebDAV base path directly on disk,
// as another user of the share would, and backdates it by age
func writeAgedDiskFile(t *testing.T, root, rel string, age time.Duration) {
	t.Helper()

	p := filepath.Join(root, "transfer", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(webdavTestContent), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", rel, err)
	}
	mtime := time.Now().Add(-age)
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatalf("Chtimes %s: %v", rel, err)
	}
}

func TestWebDAVPurgeLeavesFilesTransferShDidNotWrite(t *testing.T) {
	// The base path may be a folder on a shared drive: old files that do not
	// look like an upload must survive the purge.
	s, root := newWebDAVPurgeStorage(t, nil)
	old := 10 * 24 * time.Hour
	foreign := []string{
		"notes.txt",          // directly in the base path
		"photos/holiday.jpg", // in a directory, without metadata
		"token/sub/file.txt", // nested deeper than a token directory
		"token/sub/file.txt.metadata",
	}
	for _, rel := range foreign {
		writeAgedDiskFile(t, root, rel, old)
	}

	if err := s.Purge(context.Background(), 24*time.Hour); err != nil {
		t.Fatalf("Purge: %v", err)
	}

	for _, rel := range foreign {
		if _, err := os.Stat(filepath.Join(root, "transfer", filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s: %v, want it kept", rel, err)
		}
	}
}

func TestWebDAVPurgeDatesUploadsByTheirDataFile(t *testing.T) {
	// A download rewrites the metadata, which must not keep an old upload
	// alive.
	s, root := newWebDAVPurgeStorage(t, nil)
	putAgedWebDAVTestFile(t, s, root, "token", "file.txt", 10*24*time.Hour)
	putAgedWebDAVTestFile(t, s, root, "token", "file.txt.metadata", time.Minute)

	if err := s.Purge(context.Background(), 24*time.Hour); err != nil {
		t.Fatalf("Purge: %v", err)
	}

	for _, name := range []string{"file.txt", "file.txt.metadata"} {
		if webdavTestFileExists(t, s, "token", name) {
			t.Errorf("token/%s still exists after purge", name)
		}
	}
}
