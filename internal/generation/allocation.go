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

type allocatedObject[T any, H holder[T]] struct {
	container    H
	controllable bool
	arena        *limited_arena.LimitedArena
}

func allocate[T any, H holder[T]](gen *Generation, allocateObject allocateFunc[H, T]) allocatedObject[T, H] {
	var object allocatedObject[T, H]
	for _, arena := range gen.arenas {
		object.container, object.controllable = allocateObject(&arena)
		if object.container != nil {
			break
		}
	}
	if object.container == nil {
		arena := limited_arena.NewLimitedArena()
		object.container, object.controllable = allocateObject(&arena)
		gen.arenas = append(gen.arenas, arena)
	}

	return object
}

func AllocateObject[T any](gen *Generation) (get func() *T, finalize func()) {
	object := allocate[T](gen, limited_arena.New[T])

	metadata := objectMetadata{
		address:      unsafe.Pointer(object.container),
		typeInfo:     reflect.TypeOf(*object.container),
		controllable: object.controllable,
	}

	if !object.controllable {
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
	object := allocate[T](gen, func(arena *limited_arena.LimitedArena) ([]T, bool) {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	metadata := SliceMetadata{
		objectMetadata: objectMetadata{
			address:      unsafe.Pointer(&object.container[0]),
			typeInfo:     reflect.TypeOf(object.container),
			controllable: object.controllable,
		},
		len: len,
		cap: cap,
	}

	if !object.controllable {
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
