package main

import "github.com/Inspirate789/alloc"

func main() {
	object := alloc.New[int]()
	println(*object.Get()) // 0
	*object.Get() = 7
	println(*object.Get()) // 7
}
