package main

import (
	"arena"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

var mem *arena.Arena

// var hiddenPtr uintptr
// var hiddenPtr unsafe.Pointer

func init() {
	mem = arena.NewArena()
}

type Test struct {
	Field *TestPtr
}

type TestPtr *Test

func subWork() *TestPtr {
	ptr1 := TestPtr(arena.New[Test](mem))
	doublePtr1 := &ptr1
	runtime.SetFinalizer(doublePtr1, func(a *TestPtr) { fmt.Println("pointer finalizer 1 called") })

	ptr2 := TestPtr(arena.New[Test](mem))
	doublePtr2 := &ptr2
	runtime.SetFinalizer(doublePtr2, func(a *TestPtr) { fmt.Println("pointer finalizer 2 called") })

	// link cycle
	ptr1.Field = doublePtr2
	ptr2.Field = doublePtr1

	// add link from roots to arena
	// hiddenPtr = uintptr(unsafe.Pointer(doublePtr1))
	// hiddenPtr = unsafe.Pointer(doublePtr1)

	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	fmt.Printf("NumGC: %d\n", stats.NumGC)

	return doublePtr1
}

func Work() {
	doublePtr := subWork()
	fmt.Println((*doublePtr).Field)

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
