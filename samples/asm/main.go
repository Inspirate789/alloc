package main

import "fmt"

func getSP() int64

func logRegisters() {
	fmt.Println(getSP())
	a := make([]int, 10)
	fmt.Println(a)
	fmt.Println(getSP())
}

func main() {
	fmt.Println(getSP())
	logRegisters()
	fmt.Println(getSP())
}
