package main

func iterLStdlib(str1, str2 string) int {
	matrix := make([][]int, 0)
	for range len(str1) + 1 {
		matrix = append(matrix, make([]int, len(str2)+1))
	}

	for i := range len(str1) + 1 {
		matrix[i][0] = i
	}

	for j := range len(str2) + 1 {
		matrix[0][j] = j
	}

	for i := 1; i < len(str1)+1; i++ {
		for j := 1; j < len(str2)+1; j++ {
			var match int
			if str1[i-1] != str2[j-1] {
				match = 1
			}
			matrix[i][j] = min(
				matrix[i][j-1]+1,       // insert distance
				matrix[i-1][j]+1,       // delete distance
				matrix[i-1][j-1]+match, // match distance
			)
		}
	}

	return matrix[len(str1)][len(str2)]
}

//func main() {
//	fmt.Println(iterLStdlib("aboba", "abebaa"))
//}
