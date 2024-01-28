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

func subWork() TestPtr {
	ptr := TestPtr(arena.New[Test](mem))
	doublePtr := &ptr
	runtime.SetFinalizer(doublePtr, func(a *TestPtr) { fmt.Println("pointer finalizer called") })

	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)

	return ptr
}

func Work() {
	ptr := subWork()
	fmt.Println(ptr.A)

	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)
}

func main() {
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)

	Work()

	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)

	runtime.GC()

	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)

	time.Sleep(3 * time.Second)

	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)
}
