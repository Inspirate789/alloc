//go:build goexperiment.arenas

package limited_arena

import (
	"arena"
	"github.com/Inspirate789/alloc/internal/limited_arena/constants"
	"unsafe"
)

// LimitedArena implements a size limit on the arena.Arena
type LimitedArena struct {
	arena *arena.Arena
	free  uintptr
}

// from go/src/runtime/arena.go
const (
	// _64bit = 1 on 64-bit systems, 0 on 32-bit systems
	_64bit = 1 << (^uintptr(0) >> 63) / 2

	// heapArenaBytes is the size of a heap arena. The heap
	// consists of mappings of size heapArenaBytes, aligned to
	// heapArenaBytes. The initial heap mapping is one arena.
	//
	// This is currently 64MB on 64-bit non-Windows and 4MB on
	// 32-bit and on Windows. We use smaller arenas on Windows
	// because all committed memory is charged to the process,
	// even if it's not touched. Hence, for processes with small
	// heaps, the mapped arena space needs to be commensurate.
	// This is particularly important with the race detector,
	// since it significantly amplifies the cost of committed
	// memory.
	heapArenaBytes = 1 << logHeapArenaBytes

	// logHeapArenaBytes is log_2 of heapArenaBytes. For clarity,
	// prefer using heapArenaBytes where possible (we need the
	// constant to compute some other constants).
	logHeapArenaBytes = (6+20)*(_64bit*(1-constants.IsWindows)*(1-constants.IsWasm)*(1-constants.IsIosArm64)) + (2+20)*(_64bit*constants.IsWindows) + (2+20)*(1-_64bit) + (2+20)*constants.IsWasm + (2+20)*constants.IsIosArm64

	// userArenaChunkBytes is the size of a user arena chunk.
	userArenaChunkBytesMax = 8 << 20
	userArenaChunkBytes    = uintptr(int64(userArenaChunkBytesMax-heapArenaBytes)&(int64(userArenaChunkBytesMax-heapArenaBytes)>>63) + heapArenaBytes) // min(userArenaChunkBytesMax, heapArenaBytes)

	// userArenaChunkMaxAllocBytes is the maximum size of an object that can
	// be allocated from an arena. This number is chosen to cap worst-case
	// fragmentation of user arenas to 25%. Larger allocations are redirected
	// to the heap.
	userArenaChunkMaxAllocBytes = userArenaChunkBytes / 4

	// empirical data
	arenaSizeThreshold = userArenaChunkBytes * 95 / 100
	minArenaSize       = userArenaChunkBytes - arenaSizeThreshold
)

// NewLimitedArena allocates a new arena.
func NewLimitedArena() LimitedArena {
	return LimitedArena{
		arena: arena.NewArena(),
		free:  userArenaChunkBytes,
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

func sizeOf[T any]() uintptr {
	return unsafe.Sizeof(*new(T)) // no allocation
}

// New creates a new *T in the provided arena. The *T must not be used after
// the arena is freed. Accessing the value after free may result in a fault,
// but this fault is also not guaranteed. If there is not enough space in
// the arena, nil will be returned.
func New[T any](a *LimitedArena) (ptr *T, controllable bool) {
	size := sizeOf[T]()
	if size > userArenaChunkMaxAllocBytes {
		// WARN: the allocation will occur outside arena, i.e. in heap (from go/src/runtime/arena.go)
		return new(T), false
	} else if a.free-size >= minArenaSize {
		a.free -= size
		return arena.New[T](a.arena), true
	} else {
		return nil, false
	}
}

// MakeSlice creates a new []T with the provided capacity and length. The []T must
// not be used after the arena is freed. Accessing the underlying storage of the
// slice after free may result in a fault, but this fault is also not guaranteed.
func MakeSlice[T any](a *LimitedArena, len, cap int) (slice []T, controllable bool) {
	size := sizeOf[T]() * uintptr(cap)
	if size > userArenaChunkMaxAllocBytes {
		// WARN: the allocation will occur outside arena, i.e. in heap (from go/src/runtime/arena.go)
		return make([]T, len, cap), false
	} else if a.free-size >= minArenaSize {
		a.free -= size
		return arena.MakeSlice[T](a.arena, len, cap), true
	} else {
		return nil, false
	}
}
