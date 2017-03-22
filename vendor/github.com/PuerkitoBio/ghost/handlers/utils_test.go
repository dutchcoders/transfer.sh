package handlers

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func assertTrue(cond bool, msg string, t *testing.T) bool {
	if !cond {
		t.Error(msg)
		return false
	}
	return true
}

func assertStatus(ex, ac int, t *testing.T) {
	if ex != ac {
		t.Errorf("expected status code to be %d, got %d", ex, ac)
	}
}

func assertBody(ex []byte, res *http.Response, t *testing.T) {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if !bytes.Equal(ex, buf) {
		t.Errorf("expected body to be '%s' (%d), got '%s' (%d)", ex, len(ex), buf, len(buf))
	}
}

func assertGzippedBody(ex []byte, res *http.Response, t *testing.T) {
	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, gr)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(ex, buf.Bytes()) {
		t.Errorf("expected unzipped body to be '%s' (%d), got '%s' (%d)", ex, len(ex), buf.Bytes(), buf.Len())
	}
}

func assertHeader(hName, ex string, res *http.Response, t *testing.T) {
	hVal, ok := res.Header[hName]
	if (!ok || len(hVal) == 0) && len(ex) > 0 {
		t.Errorf("expected header %s to be %s, was not set", hName, ex)
	} else if len(hVal) > 0 && hVal[0] != ex {
		t.Errorf("expected header %s to be %s, got %s", hName, ex, hVal)
	}
}

func assertPanic(t *testing.T) {
	if err := recover(); err == nil {
		t.Error("expected a panic, got none")
	}
}
