package main

import (
	"math/rand/v2"
	"testing"
)

const matrixDim = 10

// 10 - 3		7742			7765		99.7%
// 50 - 3		222059			303197		66.6%
// 100 - 3		1495493			2195693		68.1%
// 300 - 3		52626817		78902077	66.7%
// 500 - 3		227228468		406249331	55,9%
// 750 - 3		838279720		1401996922	59,8%
// 1000 - 3		3049302628		7017907964	43,5%

var (
	resStdlib [][]int
	resAlloc  []int
)

func Rand(dim int) [][]int {
	res := make([][]int, 0, dim)
	for range dim {
		res = append(res, rand.Perm(dim))
	}

	return res
}

func BenchmarkAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resAlloc = WinogradProdAlloc(Rand(matrixDim), Rand(matrixDim))
	}
}

func BenchmarkStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resStdlib = WinogradProdStdlib(Rand(matrixDim), Rand(matrixDim))
	}
}
