package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"reflect"
	"unsafe"
)

type holder[T any] interface {
	*T | []T
}

type allocateFunc[H holder[T], T any] func(*limited_arena.Arena) (H, bool)

type allocatedObject[T any, H holder[T]] struct {
	container    H
	controllable bool
	arena        *limited_arena.Arena
}

func allocate[T any, H holder[T]](gen *Generation, allocateObject allocateFunc[H, T]) allocatedObject[T, H] {
	gen.arenasMx.Lock()
	var object allocatedObject[T, H]
	for _, arena := range gen.arenas {
		object.container, object.controllable = allocateObject(&arena)
		if object.container != nil {
			object.arena = &arena
			break
		}
	}
	if object.container == nil {
		arena := limited_arena.NewLimitedArena()
		gen.arenaSignals <- struct{}{}
		object.container, object.controllable = allocateObject(&arena)
		object.arena = &arena
		gen.arenas = append(gen.arenas, arena)
	}
	gen.arenasMx.Unlock()

	return object
}

func AllocateObject[T any](gen *Generation) *ObjectMetadata {
	object := allocate[T](gen, limited_arena.New[T])
	addr := unsafe.Pointer(object.container)

	metadata := ObjectMetadata{
		Address:  addr,
		typeInfo: reflect.TypeOf(*object.container),
		arena:    object.arena,
	}

	if !object.controllable {
		gen.uncontrollableAddresses.Add(addr, &metadata)
	} else {
		gen.addresses.Add(addr, &metadata)
	}

	return &metadata
}

func MakeSliceFromPtr[T any](ptr unsafe.Pointer, len, cap int) []T {
	return unsafe.Slice((*T)(ptr), cap)[:len]
}

func AllocateSlice[T any](gen *Generation, len, cap int) *SliceMetadata {
	object := allocate[T](gen, func(arena *limited_arena.Arena) ([]T, bool) {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})
	addr := unsafe.Pointer(&object.container[:1][0])

	metadata := SliceMetadata{
		ObjectMetadata: ObjectMetadata{
			Address:  addr,
			typeInfo: reflect.TypeOf(object.container),
			arena:    object.arena,
		},
		Len: len,
		Cap: cap,
	}

	if !object.controllable {
		gen.uncontrollableSlices.Add(addr, &metadata)
	} else {
		gen.slices.Add(addr, &metadata)
	}

	return &metadata
}

func AppendSlice[T any](gen *Generation, metadata *SliceMetadata, elems ...T) {
	metadata.Lock()
	oldAddress := metadata.Address
	slice := MakeSliceFromPtr[T](metadata.Address, metadata.Len, metadata.Cap)
	slice = append(slice, elems...) // maybe move on realloc
	newAddress := unsafe.Pointer(&slice[:1][0])
	if newAddress != oldAddress {
		gen.slices.Move(oldAddress, newAddress)
	}
	metadata.Address = newAddress
	metadata.Len = len(slice)
	metadata.Cap = cap(slice)
	metadata.Unlock()

	return
}
