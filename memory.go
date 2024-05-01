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

func allocateObject[T any](mem memory) (m *generation.ObjectMetadata) {
	size := unsafe.Sizeof(*new(T)) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateObject[T](mem.largeObjectGeneration)
	}
	return generation.AllocateObject[T](mem.movingGenerations[0])
}

func allocateSlice[T any](mem memory, len, cap int) *generation.SliceMetadata {
	size := unsafe.Sizeof(*new(T)) * uintptr(cap) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap)
	} else {
		return generation.AllocateSlice[T](mem.movingGenerations[0], len, cap)
	}
}

func appendSlice[T any](mem memory, slice []T, elems ...T) {
	slicePtr := unsafe.Pointer(&slice[:1][0])
	var gen *generation.Generation
	var metadata *generation.SliceMetadata
	var exist bool
	for _, gen = range mem.movingGenerations {
		metadata, exist = gen.SearchSliceData(slicePtr)
		if exist {
			break
		}
	}
	if !exist {
		metadata, exist = mem.largeObjectGeneration.SearchSliceData(slicePtr)
		if !exist {
			panic("slice data not found in memory")
		} else {
			gen = mem.largeObjectGeneration
		}
	}

	generation.AppendSlice(gen, metadata, elems...)
}
