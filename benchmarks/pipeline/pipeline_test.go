package main

import (
	"testing"
)

const (
	bufSize = 100_000
	count   = 10000
)

// 10 - 1		271385		320885
// 50 - 1		475905		458055
// 100 - 3		701633		678104
// 500 - 6		1811209		1719663
// 1000 - 1		3388504		3248213
// 5000 - 1		19896933	15271662
// 10000 - 1	37348724	27464640

func BenchmarkAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for range readerAlloc(writerAlloc(allocatorAlloc(bufSize, count))) {
		}
	}
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for range readerStdlib(writerStdlib(allocatorStdlib(bufSize, count))) {
		}
	}
}
