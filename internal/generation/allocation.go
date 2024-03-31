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
		address:      unsafe.Pointer(ptr),
		typeInfo:     reflect.TypeOf(*ptr),
		controllable: controllable,
	}

	if !controllable {
		gen.uncontrollableAddresses.Add(&metadata)
	} else {
		metadata.controllable = true
		gen.addresses.Add(&metadata)
	}

	get = func() *T {
		metadata.RLock()
		res := (*T)(metadata.address)
		metadata.RUnlock()
		return res
	}
	finalize = func() {
		finalized := metadata.finalized.Swap(true)
		if finalized {
			println("object is already finalized!")
		}
	}
	return
}

func makeSliceFromPtr[T any](ptr unsafe.Pointer, len, cap int) []T {
	return unsafe.Slice((*T)(ptr), cap)[:len]
}

func AllocateSlice[T any](gen *Generation, len, cap int) (get func() []T, finalize func()) {
	slice, controllable := allocate[T](gen, func(arena *limited_arena.LimitedArena) ([]T, bool) {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	metadata := SliceMetadata{
		objectMetadata: objectMetadata{
			address:      unsafe.Pointer(&slice[0]),
			typeInfo:     reflect.TypeOf(slice),
			controllable: controllable,
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
		metadata.RLock()
		res := makeSliceFromPtr[T](metadata.address, metadata.len, metadata.cap)
		metadata.RUnlock()
		return res
	}
	finalize = func() {
		finalized := metadata.finalized.Swap(true)
		if finalized {
			println("object is already finalized!")
		}
	}
	return
}

func AppendSlice[T any](metadata *SliceMetadata, elems ...T) {
	metadata.Lock()
	slice := makeSliceFromPtr[T](metadata.address, metadata.len, metadata.cap)
	slice = append(slice, elems...) // maybe move on realloc
	metadata.address = unsafe.Pointer(&slice[0])
	metadata.len = len(slice)
	metadata.cap = cap(slice)
	metadata.Unlock()

	return
}
