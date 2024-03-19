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
	slices                []*generation.SliceMetadata
}

func (mem memory) SearchSliceData(slicePtr unsafe.Pointer) *generation.SliceMetadata {
	for _, metadata := range mem.slices {
		if metadata.Address == slicePtr {
			return metadata
		}
	}
	panic("slice data not found in memory") // TODO: remove
}

func allocateObject[T any](mem memory) (get func() *T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) // no allocation
	if size > objectSizeThreshold {
		return generation.AllocateObject[T](mem.largeObjectGeneration)
	}
	return generation.AllocateObject[T](mem.youngGeneration)
}

func allocateSlice[T any](mem memory, len, cap int) (get func() []T, finalize func()) {
	size := unsafe.Sizeof(*new(T)) * uintptr(cap) // no allocation
	var data *generation.SliceMetadata
	if size > objectSizeThreshold {
		data, get, finalize = generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap)
	} else {
		data, get, finalize = generation.AllocateSlice[T](mem.youngGeneration, len, cap)
	}
	mem.slices = append(mem.slices, data)
	return
}

func appendSlice[T any](mem memory, slice []T, elems ...T) {
	metadata := mem.SearchSliceData(unsafe.Pointer(&slice[0]))
	metadata.Lock.Lock()
	generation.AppendSlice(metadata, elems...)
	metadata.Lock.Lock()
}
