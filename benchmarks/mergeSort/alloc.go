package main

import (
	"fmt"
	"github.com/Inspirate789/alloc"
)

func mergeAlloc(left []int, right []int) []int {
	leftLen, rightLen := len(left), len(right)
	res := alloc.MakeSlice[int](leftLen+rightLen, leftLen+rightLen)

	i, j := 0, 0
	for i < leftLen && j < rightLen {
		l, r := left[i], right[j]
		if l < r {
			res.Get()[i+j] = l
			i++
		} else {
			res.Get()[i+j] = r
			j++
		}
	}

	for ; i < len(left); i++ {
		res.Get()[i+j] = left[i]
	}

	for ; j < len(right); j++ {
		res.Get()[i+j] = right[j]
	}

	return res.Get()
}

func mergeSortAlloc(items []int) []int {
	if len(items) < 2 {
		return items
	}

	return mergeAlloc(
		mergeSortAlloc(items[:len(items)/2]),
		mergeSortAlloc(items[len(items)/2:]),
	)
}

func main() {
	data := []int{10, 6, 2, 1, 5, 8, 3, 4, 7, 9}
	fmt.Println(mergeSortAlloc(data))
}
