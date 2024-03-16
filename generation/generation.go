package generation

import (
	"alloc/limited_arena"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"unsafe"
)

type addressContainer[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (value V, exist bool)
	SearchByAddress(addr unsafe.Pointer) (value V, exist bool)
	Delete(key K)
}

type objectMetadata struct {
	lock       sync.Mutex
	address    unsafe.Pointer
	arenaIndex int // TODO: remove?
	finalized  atomic.Bool
}

type sliceMetadata struct {
	objectMetadata
	len int
	cap int
}

type Generation struct {
	movingObjects bool
	arenas        []limited_arena.LimitedArena
	age           uint // the amount of garbage collection that a generation has survived // TODO: remove?
	collectionMx  sync.Mutex
	addresses     addressContainer[uint64, *objectMetadata] // uuid -> metadata
	slices        addressContainer[uint64, *sliceMetadata]  // uuid -> metadata
}

type holder[T any] interface {
	*T | []T
}

func allocate[T any, H holder[T]](gen *Generation, allocateFunc func(*limited_arena.LimitedArena) H) (H, int) {
	var ptr H
	var index int
	for i, arena := range gen.arenas {
		if arena.Load() < limited_arena.MaxLoad {
			ptr = allocateFunc(&arena)
			if ptr != nil {
				index = i
				break
			}
		}
	}
	if ptr == nil {
		arena := limited_arena.NewLimitedArena()
		index = len(gen.arenas)
		ptr = allocateFunc(&arena)
		gen.arenas = append(gen.arenas, arena)
	}

	return ptr, index
}

func AllocateObject[T any](gen *Generation) (get func() *T, finalize func()) {
	ptr, arenaIndex := allocate[T](gen, limited_arena.New[T])

	metadata := objectMetadata{
		address:    unsafe.Pointer(ptr),
		arenaIndex: arenaIndex,
	}
	uuid := rand.Uint64()
	gen.addresses.Set(uuid, &metadata)

	get = func() *T {
		metadata.lock.Lock()
		res := (*T)(metadata.address)
		metadata.lock.Unlock()
		return res
	}
	finalize = func() {
		metadata.finalized.Store(true)
	}
	return
}

func makeSliceFromPtr[T any](ptr unsafe.Pointer, len, cap int) []T {
	var slice = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(ptr), len, cap}
	return *(*[]T)(unsafe.Pointer(&slice))
}

func AllocateSlice[T any](gen *Generation, len, cap int) (get func() []T, finalize func()) {
	slice, arenaIndex := allocate[T](gen, func(arena *limited_arena.LimitedArena) []T {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	metadata := sliceMetadata{
		objectMetadata: objectMetadata{
			address:    unsafe.Pointer(&slice[0]),
			arenaIndex: arenaIndex,
		},
		len: len,
		cap: cap,
	}
	uuid := rand.Uint64()
	gen.slices.Set(uuid, &metadata)

	get = func() []T {
		metadata.lock.Lock()
		res := makeSliceFromPtr[T](metadata.address, 1, 1)
		metadata.lock.Unlock()
		return res
	}
	finalize = func() {
		metadata.finalized.Store(true)
	}
	return
}

func (gen *Generation) Contains(ptr unsafe.Pointer) bool {
	_, exist := gen.addresses.SearchByAddress(ptr)
	if !exist {
		_, exist = gen.slices.SearchByAddress(ptr)
	}
	return exist
}

func (gen *Generation) Cleanup(isObservableObject func(unsafe.Pointer) bool) {
	// TODO
}
