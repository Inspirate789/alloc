package main

import (
	"arena"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

var mem *arena.Arena

func init() {
	mem = arena.NewArena()
}

type Test struct {
	A int
}

type TestPtr *Test

func subWork() *TestPtr {
	ptr := TestPtr(arena.New[Test](mem))
	doublePtr := &ptr
	runtime.SetFinalizer(doublePtr, func(a *TestPtr) { fmt.Println("pointer finalizer called") })

	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("(subWork) NumGC: %d\n", stats.NumGC)

	return doublePtr
}

func Work() {
	doublePtr := subWork()
	fmt.Println((*doublePtr).A)

	runtime.GC()
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("(Work)    NumGC: %d\n", stats.NumGC)

	go func() {
		fmt.Println("goroutine 1 started")
		doublePtr := doublePtr
		time.Sleep(10 * time.Second)
		fmt.Println(doublePtr)
		runtime.GC()
		var stats debug.GCStats
		debug.ReadGCStats(&stats)
		fmt.Printf("(g 1)     NumGC: %d\n", stats.NumGC)
		fmt.Println("goroutine 1 finished")
	}()
}

func main() {
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("(main 1)  NumGC: %d\n", stats.NumGC)

	Work()

	debug.ReadGCStats(&stats)
	fmt.Printf("(main 2)  NumGC: %d\n", stats.NumGC)

	runtime.GC()

	debug.ReadGCStats(&stats)
	fmt.Printf("(main 3)  NumGC: %d\n", stats.NumGC)

	time.Sleep(3 * time.Second)

	debug.ReadGCStats(&stats)
	fmt.Printf("(main 4)  NumGC: %d\n", stats.NumGC)
	time.Sleep(15 * time.Second)
}
