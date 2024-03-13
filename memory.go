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
	size := unsafe.Sizeof(*new(T))
	if size > objectSizeThreshold {
		return generation.AllocateObject[T](mem.largeObjectGeneration, size)
	}
	return generation.AllocateObject[T](mem.youngGeneration, size)
	// TODO: check if gc is needed ?
}

func allocateSlice[T any](mem memory, len, cap int) (get func() []T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap, size)
	}
	return generation.AllocateSlice[T](mem.youngGeneration, len, cap, size)
	// TODO: check if gc is needed ?
}
