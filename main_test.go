package main

import "testing"

func BenchmarkSimpleDownload(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SimpleDownload(url)
	}
}

func BenchmarkConcurrencyDownload(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConcurrencyDownload(url)
	}
}
