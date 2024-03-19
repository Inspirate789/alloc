package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"github.com/Inspirate789/alloc/internal/metadata_container"
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
	addr      unsafe.Pointer
	finalized atomic.Bool
}

func (om *objectMetadata) Address() unsafe.Pointer {
	return om.addr
}

type SliceMetadata struct {
	Lock      sync.Mutex
	Addr      unsafe.Pointer
	gen       *Generation
	finalized atomic.Bool
	len       int
	cap       int
}

func (sm *SliceMetadata) Address() unsafe.Pointer {
	return sm.Addr
}

type Generation struct {
	movingObjects bool
	arenas        []limited_arena.LimitedArena
	movingMx      sync.Mutex                                // must be locked at both the src and the dst
	addresses     addressContainer[uint64, *objectMetadata] // uuid -> metadata
	slices        addressContainer[uint64, *SliceMetadata]  // uuid -> metadata
}

func NewGeneration(movingObjects bool) *Generation {
	return &Generation{
		movingObjects: movingObjects,
		arenas:        []limited_arena.LimitedArena{limited_arena.NewLimitedArena()},
		addresses:     metadata_container.NewMetadataMap[uint64, *objectMetadata](0),
		slices:        metadata_container.NewMetadataMap[uint64, *SliceMetadata](0),
	}
}

func (gen *Generation) SearchSliceData(slicePtr unsafe.Pointer) (metadata *SliceMetadata, exist bool) {
	gen.movingMx.Lock()
	metadata, exist = gen.slices.SearchByAddress(slicePtr)
	gen.movingMx.Unlock()
	return
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
		addr: unsafe.Pointer(ptr),
	}
	uuid := rand.Uint64()
	gen.addresses.Set(uuid, &metadata)

	get = func() *T {
		metadata.lock.Lock()
		res := (*T)(metadata.addr)
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

func AllocateSlice[T any](gen *Generation, len, cap int) (get func() []T, finalize func()) {
	slice := allocate[T](gen, func(arena *limited_arena.LimitedArena) []T {
		return limited_arena.MakeSlice[T](arena, len, cap)
	})

	metadata := &SliceMetadata{
		Addr: unsafe.Pointer(&slice[0]),
		gen:  gen,
		len:  len,
		cap:  cap,
	}
	uuid := rand.Uint64() // TODO: is it needed?
	gen.slices.Set(uuid, metadata)

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
	slice = append(slice, elems...) // TODO: move on realloc?
	metadata.Addr = unsafe.Pointer(&slice[0])
	metadata.len = len(slice)
	metadata.cap = cap(slice)

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
