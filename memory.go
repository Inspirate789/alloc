package alloc

import (
	"alloc/generation"
	"unsafe"
)

const objectSizeThreshold = 85_000 // from dotnet

type memory struct {
	youngGeneration       *generation.Generation
	oldGeneration         *generation.Generation
	thirdGeneration       *generation.Generation
	largeObjectGeneration *generation.Generation
}

func allocateObject[T any](mem memory) (get func() *T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateObject[T](mem.largeObjectGeneration)
	}
	return generation.AllocateObject[T](mem.youngGeneration)
	// TODO: check if gc is needed ?
}

func allocateSlice[T any](mem memory, len, cap int) (get func() []T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) * uintptr(cap) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap)
	}
	return generation.AllocateSlice[T](mem.youngGeneration, len, cap)
	// TODO: check if gc is needed ?
}
