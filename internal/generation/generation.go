package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"github.com/Inspirate789/alloc/internal/metadata_container"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

type addressContainer[V any] interface {
	Add(value V)
	Search(addr unsafe.Pointer) (value V, exist bool)
	Move(old unsafe.Pointer, new unsafe.Pointer)
	Map(func(value V))
	Delete(addr unsafe.Pointer)
}

// finalized || (cyclicallyReferenced && referenceCount == 0) ==> dead object
type objectMetadata struct {
	Lock                 sync.Mutex
	Addr                 unsafe.Pointer
	typeInfo             reflect.Type
	controllable         bool
	lastMarkID           uint64
	cyclicallyReferenced bool
	referenceCount       uint // founded references (not all)
	finalized            atomic.Bool
}

func (om *objectMetadata) Address() unsafe.Pointer {
	return om.Addr // TODO: movingMx? Lock?
}

type SliceMetadata struct {
	objectMetadata
	gen *Generation // TODO: remove
	len int
	cap int
}

type Generation struct {
	movingObjects           bool
	arenas                  []limited_arena.LimitedArena
	movingMx                sync.Mutex // must be locked at both the src and the dst
	addresses               addressContainer[*objectMetadata]
	uncontrollableAddresses addressContainer[*objectMetadata]
	slices                  addressContainer[*SliceMetadata]
	uncontrollableSlices    addressContainer[*SliceMetadata]
}

func NewGeneration(movingObjects bool) *Generation {
	return &Generation{
		movingObjects:           movingObjects,
		arenas:                  []limited_arena.LimitedArena{limited_arena.NewLimitedArena()},
		addresses:               metadata_container.NewAddressMap[*objectMetadata](),
		uncontrollableAddresses: metadata_container.NewAddressMap[*objectMetadata](),
		slices:                  metadata_container.NewAddressMap[*SliceMetadata](),
		uncontrollableSlices:    metadata_container.NewAddressMap[*SliceMetadata](),
	}
}

func (gen *Generation) SearchSliceData(slicePtr unsafe.Pointer) (metadata *SliceMetadata, exist bool) {
	gen.movingMx.Lock()
	metadata, exist = gen.slices.Search(slicePtr)
	gen.movingMx.Unlock()
	return
}
