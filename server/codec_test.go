package server

import "testing"

func BenchmarkEncodeConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Token(5) + Token(5)
	}
}

func BenchmarkEncodeLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Token(10)
	}
}
