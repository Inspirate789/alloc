package main

import (
	"math/rand/v2"
	"testing"
)

const (
	n = 1_000
	k = 10_000_000
)

// 1000		10000000	14686845	18526972	79.3%
// 10000	10000000	15219658	18666888	81.5%
// 100000	10000000	16556131	20360198	81.3%
// 500000	10000000	20712474	23524767	88.0%
// 1000000	10000000	28029646	32344909	86.7%
// 5000000	10000000	79864367	78068655
// 10000000	10000000	138367079	137969266

func generate(count, max int) []int {
	data := rand.Perm(count)
	for i := range data {
		data[i] *= max / count
	}

	return data
}

func BenchmarkAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := generate(n, k)
		b.StartTimer()
		CountingSortAlloc(data, k)
	}
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := generate(n, k)
		b.StartTimer()
		CountingSortStdlib(data, k)
	}
}
