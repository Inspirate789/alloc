package alloc

import (
	"github.com/Inspirate789/alloc/internal/generation"
	"unsafe"
)

const objectSizeThreshold = 85_000 // from dotnet

type memory struct {
	movingGenerations     []*generation.Generation
	largeObjectGeneration *generation.Generation
}

func allocateObject[T any](mem memory) (get func() *T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateObject[T](mem.largeObjectGeneration)
	}
	return generation.AllocateObject[T](mem.movingGenerations[0])
}

func allocateSlice[T any](mem memory, len, cap int) (get func() []T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) * uintptr(cap) // no allocation
	if size > objectSizeThreshold {
		get, finalize = generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap)
	} else {
		get, finalize = generation.AllocateSlice[T](mem.movingGenerations[0], len, cap)
	}
	return
}

func appendSlice[T any](mem memory, slice []T, elems ...T) {
	slicePtr := unsafe.Pointer(&slice[0])
	var metadata *generation.SliceMetadata
	var exist bool
	for _, gen := range mem.movingGenerations {
		metadata, exist = gen.SearchSliceData(slicePtr)
		if exist {
			break
		}
	}
	if !exist {
		metadata, exist = mem.largeObjectGeneration.SearchSliceData(slicePtr)
		if !exist {
			panic("slice data not found in memory")
		}
	}

	generation.AppendSlice(metadata, elems...)
}
