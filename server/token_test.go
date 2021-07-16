package server

import "testing"

func BenchmarkTokenConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Token(5) + Token(5)
	}
}

func BenchmarkTokenLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Token(10)
	}
}
