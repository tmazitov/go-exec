package solutions

import (
	"testing"
)

// ---------- real test data generator ----------

func makeChannels(n, values int) []<-chan int {
	channels := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		ch := make(chan int, values)
		for j := 0; j < values; j++ {
			ch <- j
		}
		close(ch)
		channels[i] = ch
	}
	return channels
}

// ---------- benchmark helpers ----------

func benchmarkMerge(b *testing.B, fn func(...<-chan int) <-chan int, channelsCount, values int) {
	b.Helper()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		channels := makeChannels(channelsCount, values)
		out := fn(channels...)

		// drain output to ensure full execution
		for range out {
		}
	}
}

// ---------- benchmarks ----------

func BenchmarkMerge(b *testing.B) {
	benchmarks := []struct {
		name   string
		ch     int
		values int
	}{
		{"2ch_1000vals", 2, 1000},
		{"5ch_1000vals", 5, 1000},
		{"10ch_1000vals", 10, 1000},
		{"10ch_10000vals", 10, 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			benchmarkMerge(b, Merge, bm.ch, bm.values)
		})
	}
}

func BenchmarkWrongMerge(b *testing.B) {
	benchmarks := []struct {
		name   string
		ch     int
		values int
	}{
		{"2ch_1000vals", 2, 1000},
		{"5ch_1000vals", 5, 1000},
		{"10ch_1000vals", 10, 1000},
		{"10ch_10000vals", 10, 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			benchmarkMerge(b, WrongMerge, bm.ch, bm.values)
		})
	}
}
