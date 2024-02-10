package limited_arena

import (
	"arena"
	"unsafe"
)

// LimitedArena implements a size limit on the arena.Arena
type LimitedArena struct {
	arena *arena.Arena
	free  uintptr
}

const maxArenaSize = 2<<23 - 1

// NewLimitedArena allocates a new arena.
func NewLimitedArena() LimitedArena {
	return LimitedArena{
		arena: arena.NewArena(),
		free:  maxArenaSize,
	}
}

// Free frees the arena (and all objects allocated from the arena) so that
// memory backing the arena can be reused fairly quickly without garbage
// collection overhead. Applications must not call any method on this
// arena after it has been freed. If there is not enough space in the arena,
// nil will be returned.
func (a LimitedArena) Free() {
	a.arena.Free()
}

func size[T any]() uintptr {
	var val T
	return unsafe.Sizeof(val)
}

// New creates a new *T in the provided arena. The *T must not be used after
// the arena is freed. Accessing the value after free may result in a fault,
// but this fault is also not guaranteed. If there is not enough space in
// the arena, nil will be returned.
func New[T any](a *LimitedArena) *T {
	objectSize := size[T]()
	if objectSize < a.free {
		a.free -= objectSize
		return arena.New[T](a.arena)
	} else {
		return nil
	}
}

// MakeSlice creates a new []T with the provided capacity and length. The []T must
// not be used after the arena is freed. Accessing the underlying storage of the
// slice after free may result in a fault, but this fault is also not guaranteed.
func MakeSlice[T any](a *LimitedArena, len, cap int) []T {
	arraySize := size[T]() * uintptr(cap)
	if arraySize < a.free {
		a.free -= arraySize
		return arena.MakeSlice[T](a.arena, len, cap)
	} else {
		return nil
	}
}
