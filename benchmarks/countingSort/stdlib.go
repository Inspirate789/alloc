package main

func CountingSortStdlib(arr []int, k int) {
	count := make([]int, k+1)

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

//func main() {
//	data := []int{10, 6, 2, 1, 5, 8, 3, 4, 7, 9}
//	CountingSortStdlib(data)
//	fmt.Println(data)
//}
