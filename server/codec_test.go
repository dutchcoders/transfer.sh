package server

import "testing"

func BenchmarkEncodeConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Encode(INIT_SEED, 5) + Encode(INIT_SEED, 5)
	}
}

func BenchmarkEncodeLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Encode(INIT_SEED, 10)
	}
}
