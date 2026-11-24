package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
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

// A request that fails to read the file from storage must not use up a download.
func TestMissingFileDoesNotUseUpDownload(t *testing.T) {
	srvr, st := newLimitTestServer(t)
	putTestMetadata(t, st, "gone.txt", metadata{ContentType: "text/plain", MaxDownloads: 1})

	if resp := testGet(srvr, "gone.txt", ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get: status %d, want 404", resp.StatusCode)
	}
	if n := testDownloads(t, st, "gone.txt"); n != 0 {
		t.Fatalf("get: downloads %d, want 0", n)
	}

	for name, handler := range map[string]http.HandlerFunc{
		"zip":    srvr.zipHandler,
		"tar":    srvr.tarHandler,
		"tar.gz": srvr.tarGzHandler,
	} {
		req := httptest.NewRequest("GET", "/(tok/gone.txt)."+name, nil)
		req = mux.SetURLVars(req, map[string]string{"files": "tok/gone.txt"})
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d, want 404", name, w.Code)
		}
		if n := testDownloads(t, st, "gone.txt"); n != 0 {
			t.Fatalf("%s: downloads %d, want 0", name, n)
		}
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
