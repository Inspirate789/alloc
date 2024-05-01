package main

import (
	"crypto/rand"
	"testing"
)

var (
	str1 = Rand(100)
	str2 = Rand(100)
	res  int
)

// 100 - 5		184060		41720
// 300 - 3		1779617		325876
// 500 - 3		5258207		5729683
// 750 - 3		11621201	1818493
// 1000 - 5		19222116	3482899
// 3000 - 3		166382494	31616308
// 5000 - 3		544064777	86802419

func Rand(n int) string {
	buf := make([]byte, n)
	rand.Read(buf)
	return string(buf)
}

func BenchmarkAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res = iterLAlloc(str1, str2)
	}
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res = iterLStdlib(str1, str2)
	}
}
