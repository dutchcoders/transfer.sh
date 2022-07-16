package server

import "testing"

func BenchmarkTokenConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = token(5) + token(5)
	}
}

func BenchmarkTokenLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = token(10)
	}
}
