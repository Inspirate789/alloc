package main

import (
	"fmt"
	"github.com/Inspirate789/alloc"
)

func CountingSortAlloc(arr []int, k int) {
	count := alloc.MakeSlice[int](k+1, k+1).Get()

	for _, elem := range arr {
		count[elem]++
	}

	b := 0
	for i := range count {
		for range count[i] {
			arr[b] = i
			b++
		}
	}

	return
}

func main() {
	data := []int{10, 6, 2, 1, 5, 8, 3, 4, 7, 9}
	CountingSortAlloc(data, 10)
	fmt.Println(data)
}
