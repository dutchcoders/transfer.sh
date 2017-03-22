package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

type baseWriter struct{}

func (b *baseWriter) Write(data []byte) (int, error) { return 0, nil }
func (b *baseWriter) WriteHeader(code int)           {}
func (b *baseWriter) Header() http.Header            { return nil }

func TestNilWriter(t *testing.T) {
	rw, ok := GetResponseWriter(nil, func(w http.ResponseWriter) bool {
		return true
	})
	assertTrue(rw == nil, "expected nil, got non-nil", t)
	assertTrue(!ok, "expected false, got true", t)
}

func TestBaseWriter(t *testing.T) {
	bw := &baseWriter{}
	rw, ok := GetResponseWriter(bw, func(w http.ResponseWriter) bool {
		return true
	})
	assertTrue(rw == bw, fmt.Sprintf("expected %#v, got %#v", bw, rw), t)
	assertTrue(ok, "expected true, got false", t)
}

func TestWrappedWriter(t *testing.T) {
	bw := &baseWriter{}
	ctx := &contextResponseWriter{bw, nil}
	rw, ok := GetResponseWriter(ctx, func(w http.ResponseWriter) bool {
		_, ok := w.(*baseWriter)
		return ok
	})
	assertTrue(rw == bw, fmt.Sprintf("expected %#v, got %#v", bw, rw), t)
	assertTrue(ok, "expected true, got false", t)
}

func TestWrappedNotFoundWriter(t *testing.T) {
	bw := &baseWriter{}
	ctx := &contextResponseWriter{bw, nil}
	rw, ok := GetResponseWriter(ctx, func(w http.ResponseWriter) bool {
		_, ok := w.(*statusResponseWriter)
		return ok
	})
	assertTrue(rw == nil, fmt.Sprintf("expected nil, got %#v", rw), t)
	assertTrue(!ok, "expected false, got true", t)
}
