package main

import (
	"fmt"
	"runtime"
	"time"
)

type Test struct {
	A int
}

func subWork() *Test {
	a := new(Test)
	runtime.SetFinalizer(a, func(a *Test) { fmt.Println("finalizer called") })
	return a
}

func Work() {
	a := subWork()
	fmt.Println(a.A)
}

func main() {
	Work()
	runtime.GC()
	time.Sleep(3 * time.Second)
}
