package main

func prodRowsStdlib(matrix [][]int) []int {
	res := make([]int, 0)
	for i := range matrix {
		val := 0
		for j := range len(matrix[0]) / 2 {
			val += matrix[i][2*j+1] * matrix[i][2*j]
		}
		res = append(res, val)
	}

	return res
}

func prodColsStdlib(matrix [][]int) []int {
	res := make([]int, 0)
	for i := range matrix[0] {
		val := 0
		for j := range len(matrix) / 2 {
			val += matrix[2*j+1][i] * matrix[2*j][i]
		}
		res = append(res, val)
	}

	return res
}

func WinogradProdStdlib(matrix1, matrix2 [][]int) [][]int {
	res := make([][]int, 0)

	prod1 := prodRowsStdlib(matrix1)
	prod2 := prodColsStdlib(matrix2)

	for i := range matrix1 {
		row := make([]int, 0)
		for j := range matrix2[0] {
			value := -prod1[i] - prod2[j]

			for k := range len(matrix2) / 2 {
				value += (matrix1[i][2*k] + matrix2[2*k+1][j]) * (matrix1[i][2*k+1] + matrix2[2*k][j])
			}

			if len(matrix2)%2 != 0 {
				value += matrix1[i][len(matrix2)-1] * matrix2[len(matrix2)-1][j]
			}

			row = append(row, value)
		}
		res = append(res, row)
	}

	return res
}

//func main() {
//	matrix1 := [][]int{
//		{1, 2, 3},
//		{4, 5, 6},
//		{7, 8, 9},
//	}
//	matrix2 := [][]int{
//		{1, 0, 0},
//		{0, 1, 0},
//		{0, 0, 1},
//	}
//
//	res := WinogradProdStdlib(matrix1, matrix2)
//
//	for i := range res {
//		fmt.Println(res[i])
//	}
//}
