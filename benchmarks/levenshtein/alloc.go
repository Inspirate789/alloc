package main

import (
	"fmt"
	"github.com/Inspirate789/alloc"
)

func iterLAlloc(str1, str2 string) int {
	matrix := alloc.MakeSlice[int]((len(str1)+1)*(len(str2)+1), (len(str1)+1)*(len(str2)+1))

	for i := range len(str1) + 1 {
		matrix.Get()[i*(len(str1)+1)] = i
	}

	for j := range len(str2) + 1 {
		matrix.Get()[j] = j
	}

	for i := 1; i < len(str1)+1; i++ {
		for j := 1; j < len(str2)+1; j++ {
			var match int
			if str1[i-1] != str2[j-1] {
				match = 1
			}
			matrix.Get()[i*(len(str1)+1)+j] = min(
				matrix.Get()[i*(len(str1)+1)+j-1]+1,         // insert distance
				matrix.Get()[(i-1)*(len(str1)+1)+j]+1,       // delete distance
				matrix.Get()[(i-1)*(len(str1)+1)+j-1]+match, // match distance
			)
		}
	}

	return matrix.Get()[len(str1)*(len(str1)+1)+len(str2)]
}

func main() {
	fmt.Println(iterLAlloc("aboba", "abebaa"))
}
