package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"github.com/Inspirate789/alloc/internal/metadata_container"
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
