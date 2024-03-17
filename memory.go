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
	slices                []generation.SliceData
}

func (mem memory) SearchSliceData(slicePtr uintptr) generation.SliceData {
	for _, data := range mem.slices {
		if data.Address.Load() == slicePtr {
			return data
		}
	}
	panic("slice data not found in memory")
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
	var data generation.SliceData
	if size > objectSizeThreshold {
		data, get, finalize = generation.AllocateSlice[T](mem.largeObjectGeneration, len, cap)
	} else {
		data, get, finalize = generation.AllocateSlice[T](mem.youngGeneration, len, cap)
	}
	mem.slices = append(mem.slices, data)
	return
}

func appendSlice[T any](mem memory, slice []T, elems ...T) {
	data := mem.SearchSliceData(uintptr(unsafe.Pointer(&slice[0])))
	gen := data.Gen.Load()
	generation.AppendSlice(gen, slice, elems...)
}
