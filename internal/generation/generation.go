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

type gcMetadata struct {
	lastMarkID           uint64
	cyclicallyReferenced bool
	referenceCount       uint // founded references (not all) // TODO: drop to 0 on start new collection (on mark when object with old lastMarkID appears for the first time)
	finalized            atomic.Bool
	arena                *limited_arena.Arena
}

type objectMetadata struct {
	sync.RWMutex
	address  unsafe.Pointer
	typeInfo reflect.Type
	gcMetadata
}

func (om *objectMetadata) Address() (addr unsafe.Pointer) {
	om.RLock()
	addr = om.address
	om.RUnlock()
	return
}

type SliceMetadata struct {
	objectMetadata
	len int
	cap int
}

type Generation struct {
	arenas                  []limited_arena.Arena
	addresses               addressContainer[*objectMetadata]
	uncontrollableAddresses addressContainer[*objectMetadata]
	slices                  addressContainer[*SliceMetadata]
	uncontrollableSlices    addressContainer[*SliceMetadata]
}

func NewGeneration() *Generation {
	return &Generation{
		arenas:                  []limited_arena.Arena{limited_arena.NewLimitedArena()},
		addresses:               metadata_container.NewAddressMap[*objectMetadata](),
		uncontrollableAddresses: metadata_container.NewAddressMap[*objectMetadata](),
		slices:                  metadata_container.NewAddressMap[*SliceMetadata](),
		uncontrollableSlices:    metadata_container.NewAddressMap[*SliceMetadata](),
	}
}

func (gen *Generation) SearchSliceData(slicePtr unsafe.Pointer) (metadata *SliceMetadata, exist bool) {
	metadata, exist = gen.slices.Search(slicePtr)
	return
}
