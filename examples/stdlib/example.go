package main

func main() {
	object := new(int)
	println(*object) // 0
	*object = 7
	println(*object) // 7
}
