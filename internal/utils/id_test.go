package utils

import "testing"

func BenchmarkID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = UniqueSecureID60()
	}
}
