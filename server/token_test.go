package server

import (
	"io"
	"log"
	"testing"
)

var logger = log.New(io.Discard, "", log.LstdFlags)

func BenchmarkTokenConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = token(5, logger) + token(5, logger)
	}
}

func BenchmarkTokenLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = token(10, logger)
	}
}
