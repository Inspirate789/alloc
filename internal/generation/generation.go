package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"sync"
	"sync/atomic"
	"unsafe"
)

type addressContainer[V any] interface {
	Add(value V)
	SearchByAddress(addr unsafe.Pointer) (value V, exist bool)
	Delete(addr unsafe.Pointer)
}

// finalized || (cyclicallyReferenced && referenceCount == 0) ==> dead object
type gcFields struct {
	cyclicallyReferenced bool
	referenceCount       uint // founded references (not all)
	finalized            atomic.Bool
}

type objectMetadata struct {
	lock sync.Mutex
	addr unsafe.Pointer
	gcFields
}

func (om *objectMetadata) Address() unsafe.Pointer {
	return om.addr
}

type SliceMetadata struct {
	Lock sync.Mutex
	Addr unsafe.Pointer
	gen  *Generation
	len  int
	cap  int
	gcFields
}

func (sm *SliceMetadata) Address() unsafe.Pointer {
	return sm.Addr
}

type Generation struct {
	movingObjects bool
	arenas        []limited_arena.LimitedArena
	movingMx      sync.Mutex                        // must be locked at both the src and the dst
	addresses     addressContainer[*objectMetadata] // uuid -> metadata
	slices        addressContainer[*SliceMetadata]  // uuid -> metadata
}

func NewGeneration(movingObjects bool) *Generation {
	return &Generation{
		movingObjects: movingObjects,
		arenas:        []limited_arena.LimitedArena{limited_arena.NewLimitedArena()},
		addresses:     nil, // TODO
		slices:        nil, // TODO
	}
}

func (gen *Generation) SearchSliceData(slicePtr unsafe.Pointer) (metadata *SliceMetadata, exist bool) {
	gen.movingMx.Lock()
	metadata, exist = gen.slices.SearchByAddress(slicePtr)
	gen.movingMx.Unlock()
	return
}
