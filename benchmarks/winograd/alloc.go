package main

import (
	"fmt"
	"github.com/Inspirate789/alloc"
)

func prodRowsAlloc(matrix [][]int) alloc.SliceGetter[int] {
	res := alloc.MakeSlice[int](len(matrix), len(matrix))
	for i := range matrix {
		val := 0
		for j := range len(matrix[0]) / 2 {
			val += matrix[i][2*j+1] * matrix[i][2*j]
		}
		res.Get()[i] = val
	}

	return res
}

func prodColsAlloc(matrix [][]int) alloc.SliceGetter[int] {
	res := alloc.MakeSlice[int](len(matrix), len(matrix))
	for i := range matrix[0] {
		val := 0
		for j := range len(matrix) / 2 {
			val += matrix[2*j+1][i] * matrix[2*j][i]
		}
		res.Get()[i] = val
	}

	return res
}

func WinogradProdAlloc(matrix1, matrix2 [][]int) []int {
	res := alloc.MakeSlice[int](len(matrix1)*len(matrix2[0]), len(matrix1)*len(matrix2[0]))

	prod1 := prodRowsAlloc(matrix1)
	prod2 := prodColsAlloc(matrix2)

	for i := range matrix1 {
		for j := range matrix2[0] {
			value := -prod1.Get()[i] - prod2.Get()[j]

			for k := range len(matrix2) / 2 {
				value += (matrix1[i][2*k] + matrix2[2*k+1][j]) * (matrix1[i][2*k+1] + matrix2[2*k][j])
			}

			if len(matrix2)%2 != 0 {
				value += matrix1[i][len(matrix2)-1] * matrix2[len(matrix2)-1][j]
			}

			res.Get()[i*len(matrix2[0])+j] = value
		}
	}

	return res.Get()
}

func main() {
	matrix1 := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	matrix2 := [][]int{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	res := WinogradProdAlloc(matrix1, matrix2)

	for i := range res {
		fmt.Println(res[i])
	}
}
