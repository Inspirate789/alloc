package main

import (
	"arena"
	"fmt"
	"runtime"
	"time"
)

var mem *arena.Arena

func init() {
	mem = arena.NewArena()
}

type Test struct {
	A int
}

func subWork() *[]Test {
	buf := arena.MakeSlice[Test](mem, 7, 7)
	runtime.SetFinalizer(&buf, func(a *[]Test) { fmt.Println("finalizer called") })
	return &buf
}

func Work() {
	bufPtr := subWork()
	fmt.Println((*bufPtr)[0].A)
}

func main() {
	Work()
	runtime.GC()
	time.Sleep(3 * time.Second)
}
