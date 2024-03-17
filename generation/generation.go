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
	lock      sync.Mutex
	address   unsafe.Pointer
	finalized atomic.Bool
}

func (om *objectMetadata) Address() unsafe.Pointer {
	return om.address
}

type sliceMetadata struct {
	lock      sync.Mutex
	address   *atomic.Uintptr
	finalized atomic.Bool
	gen       *atomic.Pointer[Generation]
	len       int
	cap       int
}

type SliceData struct {
	Address *atomic.Uintptr // unsafe.Pointer
	Gen     *atomic.Pointer[Generation]
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

func allocate[T any, H holder[T]](gen *Generation, allocateFunc func(*limited_arena.LimitedArena) H) H {
	var ptr H
	for _, arena := range gen.arenas {
		if arena.Load() < limited_arena.MaxLoad {
			ptr = allocateFunc(&arena)
			if ptr != nil {
				break
			}
		}
	}
	if ptr == nil {
		arena := limited_arena.NewLimitedArena()
		ptr = allocateFunc(&arena)
		gen.arenas = append(gen.arenas, arena)
	}

	return ptr
}

func AllocateObject[T any](gen *Generation) (get func() *T, finalize func()) {
	ptr := allocate[T](gen, limited_arena.New[T])

	metadata := objectMetadata{
		address: unsafe.Pointer(ptr),
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

func makeSliceFromPtr[T any](ptr uintptr, len, cap int) []T {
	var slice = struct {
		addr uintptr
		len  int
		cap  int
	}{ptr, len, cap}
	return *(*[]T)(unsafe.Pointer(&slice))
}

func AllocateSlice[T any](gen *Generation, len, cap int) (data SliceData, get func() []T, finalize func()) {
	slice := allocate[T](gen, func(arena *limited_arena.LimitedArena) []T {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	var genPtr *atomic.Pointer[Generation]
	genPtr.Store(gen)
	var address *atomic.Uintptr
	address.Store(uintptr(unsafe.Pointer(&slice[0])))
	metadata := sliceMetadata{
		address: address,
		gen:     genPtr,
		len:     len,
		cap:     cap,
	}
	uuid := rand.Uint64() // TODO: is it needed?
	gen.slices.Set(uuid, &metadata)

	data = SliceData{
		Address: nil,
		Gen:     metadata.gen,
	}
	get = func() []T {
		metadata.lock.Lock()
		res := makeSliceFromPtr[T](metadata.address.Load(), metadata.len, metadata.cap)
		metadata.lock.Unlock()
		return res
	}
	finalize = func() {
		metadata.finalized.Store(true)
	}
	return
}

func AppendSlice[T any](gen *Generation, slice []T, elems ...T) {
	metadata, exist := gen.slices.SearchByAddress(unsafe.Pointer(&slice[0]))
	if !exist {
		panic("slice not found in generation") // TODO: remove
	}

	metadata.lock.Lock()
	slice = append(slice, elems...)
	metadata.address.Store(uintptr(unsafe.Pointer(&slice[0])))
	metadata.len = len(slice)
	metadata.cap = cap(slice)
	metadata.lock.Unlock()

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
