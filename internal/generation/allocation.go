package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"reflect"
	"unsafe"
)

type holder[T any] interface {
	*T | []T
}

type allocateFunc[H holder[T], T any] func(*limited_arena.LimitedArena) (H, bool)

func allocate[T any, H holder[T]](gen *Generation, allocateObject allocateFunc[H, T]) (ptr H, controllable bool) {
	for _, arena := range gen.arenas {
		ptr, controllable = allocateObject(&arena)
		if ptr != nil {
			break
		}
	}
	if ptr == nil {
		arena := limited_arena.NewLimitedArena()
		ptr, controllable = allocateObject(&arena)
		gen.arenas = append(gen.arenas, arena)
	}

	return ptr, controllable
}

func AllocateObject[T any](gen *Generation) (get func() *T, finalize func()) {
	ptr, controllable := allocate[T](gen, limited_arena.New[T])

	metadata := objectMetadata{
		Addr:     unsafe.Pointer(ptr),
		typeInfo: reflect.TypeOf(*ptr),
	}

	if !controllable {
		gen.uncontrollableAddresses.Add(&metadata)
	} else {
		metadata.controllable = true
		gen.addresses.Add(&metadata)
	}

	get = func() *T {
		metadata.Lock.Lock()
		res := (*T)(metadata.Addr)
		metadata.Lock.Unlock()
		return res
	}
	finalize = func() {
		metadata.finalized.Store(true)
	}
	return
}

func makeSliceFromPtr[T any](ptr uintptr, len, cap int) []T {
	var slice = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, cap}
	return *(*[]T)(unsafe.Pointer(&slice))
}

func AllocateSlice[T any](gen *Generation, len, cap int) (get func() []T, finalize func()) {
	slice, controllable := allocate[T](gen, func(arena *limited_arena.LimitedArena) ([]T, bool) {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	metadata := SliceMetadata{
		objectMetadata: objectMetadata{
			Addr:     unsafe.Pointer(&slice[0]),
			typeInfo: reflect.TypeOf(slice),
		},
		len: len,
		cap: cap,
	}

	if !controllable {
		gen.uncontrollableSlices.Add(&metadata)
	} else {
		metadata.controllable = true
		gen.slices.Add(&metadata)
	}

	get = func() []T {
		metadata.Lock.Lock()
		res := makeSliceFromPtr[T](uintptr(metadata.Addr), metadata.len, metadata.cap)
		metadata.Lock.Unlock()
		return res
	}
	finalize = func() {
		metadata.finalized.Store(true)
	}
	return
}

func AppendSlice[T any](metadata *SliceMetadata, elems ...T) {
	slice := makeSliceFromPtr[T](uintptr(metadata.Addr), metadata.len, metadata.cap)
	slice = append(slice, elems...) // maybe move on realloc
	metadata.Addr = unsafe.Pointer(&slice[0])
	metadata.len = len(slice)
	metadata.cap = cap(slice)

	return
}
