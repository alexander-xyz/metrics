package main

import (
	"context"
	"testing"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func BenchmarkCollectMetrics(b *testing.B) {
	store := repository.NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collectMetrics(ctx, store)
	}
}

func BenchmarkCompress(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte('a' + i%26)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compress(data)
	}
}
