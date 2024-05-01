package main

import (
	"math/rand/v2"
	"testing"
)

var (
	data = rand.Perm(500)
	res  []int
)

// 100 - 4		91084		13046
// 500 - 3		693611		110748
// 1000 - 3		992665		240706
// 5000 - 4		5137453		1565943
// 10000 - 4	10036620	3854623
// 50000 - 3	51518979	16459040
// 100000 - 4	101879139	42221259

func BenchmarkAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res = mergeSortAlloc(data)
	}
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res = mergeSortStdlib(data)
	}
}
