package main

func mergeStdlib(left []int, right []int) (res []int) {
	res = make([]int, 0) // make([]int, 0, len(left)+len(right))

	i := 0
	j := 0
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			res = append(res, left[i])
			i++
		} else {
			res = append(res, right[j])
			j++
		}
	}

	for ; i < len(left); i++ {
		res = append(res, left[i])
	}

	for ; j < len(right); j++ {
		res = append(res, right[j])
	}

	return res
}

func mergeSortStdlib(items []int) []int {
	if len(items) < 2 {
		return items
	}

	return mergeStdlib(
		mergeSortStdlib(items[:len(items)/2]),
		mergeSortStdlib(items[len(items)/2:]),
	)
}

//func main() {
//	data := []int{10, 6, 2, 1, 5, 8, 3, 4, 7, 9}
//	fmt.Println(mergeSortStdlib(data))
//}
