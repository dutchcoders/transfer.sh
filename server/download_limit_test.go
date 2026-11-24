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

// A request that fails to decrypt must not use up a download.
func TestWrongPasswordDoesNotUseUpDownload(t *testing.T) {
	ctx := context.Background()
	logger := log.New(io.Discard, "", 0)

	st, err := storage.NewLocalStorage(t.TempDir(), logger)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("hello secret")
	encrypted, err := attachEncryptionReader(io.NopCloser(bytes.NewReader(plaintext)), "pw")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := io.ReadAll(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Put(ctx, "tok", "secret.txt", bytes.NewReader(ciphertext), "text/plain", uint64(len(ciphertext))); err != nil {
		t.Fatal(err)
	}

	meta, err := json.Marshal(metadata{
		ContentType:          "text/plain",
		ContentLength:        int64(len(plaintext)),
		MaxDownloads:         1,
		Encrypted:            true,
		DecryptedContentType: "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Put(ctx, "tok", "secret.txt.metadata", bytes.NewReader(meta), "text/json", uint64(len(meta))); err != nil {
		t.Fatal(err)
	}

	srvr, err := New(UseStorage(st), Logger(logger))
	if err != nil {
		t.Fatal(err)
	}

	get := func(password string) *http.Response {
		req := httptest.NewRequest("GET", "/tok/secret.txt", nil)
		req.Header.Set("X-Decrypt-Password", password)
		req = mux.SetURLVars(req, map[string]string{"token": "tok", "filename": "secret.txt"})
		w := httptest.NewRecorder()
		srvr.getHandler(w, req)
		return w.Result()
	}

	if resp := get("typo"); resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("wrong password: status %d, want 500", resp.StatusCode)
	}

	resp := get("pw")
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !bytes.Equal(body, plaintext) {
		t.Fatalf("right password: status %d, body %q", resp.StatusCode, body)
	}

	if resp := get("pw"); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("after the last download: status %d, want 404", resp.StatusCode)
	}
}
