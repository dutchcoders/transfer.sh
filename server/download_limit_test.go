package server

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dutchcoders/transfer.sh/server/storage"
	"github.com/gorilla/mux"
)

func newLimitTestServer(t *testing.T) (*Server, storage.Storage) {
	t.Helper()
	logger := log.New(io.Discard, "", 0)

	st, err := storage.NewLocalStorage(t.TempDir(), logger)
	if err != nil {
		t.Fatal(err)
	}

	srvr, err := New(UseStorage(st), Logger(logger))
	if err != nil {
		t.Fatal(err)
	}

	return srvr, st
}

func putTestFile(t *testing.T, st storage.Storage, filename string, content []byte) {
	t.Helper()
	if err := st.Put(context.Background(), "tok", filename, bytes.NewReader(content), "text/plain", uint64(len(content))); err != nil {
		t.Fatal(err)
	}
}

func putTestMetadata(t *testing.T, st storage.Storage, filename string, meta metadata) {
	t.Helper()
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	putTestFile(t, st, filename+".metadata", data)
}

func testDownloads(t *testing.T, st storage.Storage, filename string) int {
	t.Helper()
	r, _, err := st.Get(context.Background(), "tok", filename+".metadata", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.CloseCheck(r)

	var meta metadata
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	return meta.Downloads
}

func testGet(srvr *Server, filename, password string) *http.Response {
	req := httptest.NewRequest("GET", "/tok/"+filename, nil)
	if password != "" {
		req.Header.Set("X-Decrypt-Password", password)
	}
	req = mux.SetURLVars(req, map[string]string{"token": "tok", "filename": filename})
	w := httptest.NewRecorder()
	srvr.getHandler(w, req)
	return w.Result()
}

// A request that fails to decrypt must not use up a download.
func TestWrongPasswordDoesNotUseUpDownload(t *testing.T) {
	srvr, st := newLimitTestServer(t)

	plaintext := []byte("hello secret")
	encrypted, err := attachEncryptionReader(io.NopCloser(bytes.NewReader(plaintext)), "pw")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	putTestFile(t, st, "secret.txt", ciphertext)
	putTestMetadata(t, st, "secret.txt", metadata{
		ContentType:          "text/plain",
		ContentLength:        int64(len(plaintext)),
		MaxDownloads:         1,
		Encrypted:            true,
		DecryptedContentType: "text/plain",
	})

	if resp := testGet(srvr, "secret.txt", "typo"); resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("wrong password: status %d, want 500", resp.StatusCode)
	}
	if n := testDownloads(t, st, "secret.txt"); n != 0 {
		t.Fatalf("wrong password: downloads %d, want 0", n)
	}

	resp := testGet(srvr, "secret.txt", "pw")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, plaintext) {
		t.Fatalf("right password: status %d, body %q", resp.StatusCode, body)
	}

	if resp := testGet(srvr, "secret.txt", "pw"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("after the last download: status %d, want 404", resp.StatusCode)
	}
}

// An archive skips the requested files that do not exist (it is empty when
// none do) and counts a download only for the files it contains.
func TestArchiveCountsOnlyServedFiles(t *testing.T) {
	for _, existing := range [][]string{nil, {"a.txt"}, {"a.txt", "b.txt"}} {
		for _, ext := range []string{"zip", "tar", "tar.gz"} {
			srvr, st := newLimitTestServer(t)
			handler := map[string]http.HandlerFunc{
				"zip":    srvr.zipHandler,
				"tar":    srvr.tarHandler,
				"tar.gz": srvr.tarGzHandler,
			}[ext]
			for _, name := range existing {
				putTestFile(t, st, name, []byte(name))
				putTestMetadata(t, st, name, metadata{ContentType: "text/plain", MaxDownloads: 2})
			}
			// gone.txt has metadata but its data is gone, nofile.txt was never uploaded
			putTestMetadata(t, st, "gone.txt", metadata{ContentType: "text/plain", MaxDownloads: 2})

			var keys []string
			for _, name := range append(existing, "gone.txt", "nofile.txt") {
				keys = append(keys, "tok/"+name)
			}
			files := strings.Join(keys, ",")
			req := httptest.NewRequest("GET", "/("+files+")."+ext, nil)
			req = mux.SetURLVars(req, map[string]string{"files": files})
			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("%s %v: status %d, want 200", ext, existing, w.Code)
			}
			if got := archiveNames(t, ext, w.Body.Bytes()); strings.Join(got, ",") != strings.Join(existing, ",") {
				t.Fatalf("%s %v: archive has %v", ext, existing, got)
			}
			for _, name := range existing {
				if n := testDownloads(t, st, name); n != 1 {
					t.Fatalf("%s %v: %s downloads %d, want 1", ext, existing, name, n)
				}
			}
		}
	}
}

func archiveNames(t *testing.T, ext string, body []byte) []string {
	t.Helper()
	var names []string
	if ext == "zip" {
		zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range zr.File {
			names = append(names, f.Name)
		}
		return names
	}

	var r io.Reader = bytes.NewReader(body)
	if ext == "tar.gz" {
		gr, err := gzip.NewReader(r)
		if err != nil {
			t.Fatal(err)
		}
		r = gr
	}
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return names
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, h.Name)
	}
}

// A successful download is counted exactly once.
func TestSuccessfulDownloadCountsOnce(t *testing.T) {
	srvr, st := newLimitTestServer(t)
	content := []byte("hello")
	putTestFile(t, st, "hello.txt", content)
	putTestMetadata(t, st, "hello.txt", metadata{ContentType: "text/plain", MaxDownloads: 2})

	resp := testGet(srvr, "hello.txt", "")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, content) {
		t.Fatalf("status %d, body %q", resp.StatusCode, body)
	}
	if n := testDownloads(t, st, "hello.txt"); n != 1 {
		t.Fatalf("downloads %d, want 1", n)
	}
}

// Without a password an encrypted file is served as stored, and that counts.
func TestEncryptedWithoutPasswordCounts(t *testing.T) {
	srvr, st := newLimitTestServer(t)

	encrypted, err := attachEncryptionReader(io.NopCloser(bytes.NewReader([]byte("hello secret"))), "pw")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	putTestFile(t, st, "secret.txt", ciphertext)
	putTestMetadata(t, st, "secret.txt", metadata{ContentType: "text/plain", MaxDownloads: 2, Encrypted: true})

	resp := testGet(srvr, "secret.txt", "")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, ciphertext) {
		t.Fatalf("status %d, body %q", resp.StatusCode, body)
	}
	if n := testDownloads(t, st, "secret.txt"); n != 1 {
		t.Fatalf("downloads %d, want 1", n)
	}
}

// Every Range request counts as a download.
func TestRangeRequestCounts(t *testing.T) {
	srvr, st := newLimitTestServer(t)
	putTestFile(t, st, "hello.txt", []byte("hello"))
	putTestMetadata(t, st, "hello.txt", metadata{ContentType: "text/plain", MaxDownloads: 2})

	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/tok/hello.txt", nil)
		req.Header.Set("Range", "bytes=1-2")
		req = mux.SetURLVars(req, map[string]string{"token": "tok", "filename": "hello.txt"})
		w := httptest.NewRecorder()
		srvr.getHandler(w, req)

		if w.Code != http.StatusPartialContent || w.Body.String() != "el" {
			t.Fatalf("range %d: status %d, body %q", i, w.Code, w.Body.String())
		}
		if n := testDownloads(t, st, "hello.txt"); n != i {
			t.Fatalf("range %d: downloads %d, want %d", i, n, i)
		}
	}

	if resp := testGet(srvr, "hello.txt", ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("after the last download: status %d, want 404", resp.StatusCode)
	}
}

// A request answered with an error status does not count.
func TestErrorStatusDoesNotCount(t *testing.T) {
	srvr, st := newLimitTestServer(t)
	putTestFile(t, st, "hello.txt", []byte("hello"))
	putTestMetadata(t, st, "hello.txt", metadata{ContentType: "text/plain", MaxDownloads: 1})

	if resp := testGet(srvr, "hello.txt", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("first download: status %d, want 200", resp.StatusCode)
	}
	for i := 0; i < 2; i++ {
		if resp := testGet(srvr, "hello.txt", ""); resp.StatusCode != http.StatusNotFound {
			t.Fatalf("over the limit: status %d, want 404", resp.StatusCode)
		}
	}
	if n := testDownloads(t, st, "hello.txt"); n != 1 {
		t.Fatalf("downloads %d, want 1", n)
	}
}
