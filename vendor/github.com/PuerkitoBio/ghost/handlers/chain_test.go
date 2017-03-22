package handlers

import (
	"bytes"
	"net/http"
	"testing"
)

func TestChaining(t *testing.T) {
	var buf bytes.Buffer

	a := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('a')
	}
	b := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('b')
	}
	c := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('c')
	}
	f := NewChainableHandler(http.HandlerFunc(a)).Chain(http.HandlerFunc(b)).Chain(http.HandlerFunc(c))
	f.ServeHTTP(nil, nil)

	if buf.String() != "abc" {
		t.Errorf("expected 'abc', got %s", buf.String())
	}
}

func TestChainingWithHelperFunc(t *testing.T) {
	var buf bytes.Buffer

	a := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('a')
	})
	b := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('b')
	})
	c := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('c')
	})
	d := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('d')
	})
	f := ChainHandlers(a, b, c, d)
	f.ServeHTTP(nil, nil)

	if buf.String() != "abcd" {
		t.Errorf("expected 'abcd', got %s", buf.String())
	}
}

func TestChainingMixed(t *testing.T) {
	var buf bytes.Buffer

	a := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('a')
	})
	b := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('b')
	})
	c := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('c')
	})
	d := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('d')
	})
	f := NewChainableHandler(a).Chain(ChainHandlers(b, c)).Chain(d)
	f.ServeHTTP(nil, nil)

	if buf.String() != "abcd" {
		t.Errorf("expected 'abcd', got %s", buf.String())
	}
}
