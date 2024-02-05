package main

import (
	"fmt"
	"runtime"
	"time"
)

type Test struct {
	A int
}

func subWork() []Test {
	buf := make([]Test, 7)
	runtime.SetFinalizer(&buf[0], func(a *Test) { fmt.Println("finalizer called") })
	// runtime.SetFinalizer(&buf[0], func(a *Test) { fmt.Println("finalizer 2 called") }) // fatal error: runtime.SetFinalizer: finalizer already set
	return buf
}

func Work() {
	buf := subWork()
	fmt.Println(buf[0].A)
}

func main() {
	Work()
	runtime.GC()
	time.Sleep(3 * time.Second)
}
