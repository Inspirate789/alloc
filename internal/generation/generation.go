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
	Map(func(value V))
	MoveTo(container any)
	Delete(addresses []unsafe.Pointer)
}

type gcMetadata struct {
	lastMarkID              uint64
	cycleReferenceSource    *objectMetadata
	cycleReferenceCompleted bool
	referenceCount          int // founded references (not all)
	finalized               atomic.Bool
	arena                   *limited_arena.Arena
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
	arenasMx                sync.RWMutex
	arenas                  []limited_arena.Arena
	arenaSignals            chan<- struct{}
	addresses               addressContainer[*objectMetadata]
	uncontrollableAddresses addressContainer[*objectMetadata]
	slices                  addressContainer[*SliceMetadata]
	uncontrollableSlices    addressContainer[*SliceMetadata]
}

func NewGeneration(arenaSignals chan<- struct{}) *Generation {
	return &Generation{
		arenas:                  []limited_arena.Arena{limited_arena.NewLimitedArena()},
		arenaSignals:            arenaSignals,
		addresses:               metadata_container.NewAddressContainer[*objectMetadata](),
		uncontrollableAddresses: metadata_container.NewAddressContainer[*objectMetadata](),
		slices:                  metadata_container.NewAddressContainer[*SliceMetadata](),
		uncontrollableSlices:    metadata_container.NewAddressContainer[*SliceMetadata](),
	}
}

func (gen *Generation) SearchSliceData(slicePtr unsafe.Pointer) (metadata *SliceMetadata, exist bool) {
	metadata, exist = gen.slices.Search(slicePtr)
	return
}

func (gen *Generation) Size() int {
	gen.arenasMx.RLock()
	size := len(gen.arenas)
	gen.arenasMx.RUnlock()
	return size
}

func (gen *Generation) MoveTo(dst *Generation) {
	gen.arenasMx.Lock()
	dst.arenasMx.Lock()

	gen.arenas[len(gen.arenas)-1], gen.arenas[0] = gen.arenas[0], gen.arenas[len(gen.arenas)-1]
	dst.arenas = append(dst.arenas, gen.arenas[1:]...)
	gen.arenas = gen.arenas[:1]

	dst.arenasMx.Unlock()
	gen.arenasMx.Unlock()

	gen.addresses.MoveTo(dst.addresses)
	gen.slices.MoveTo(dst.slices)
	gen.uncontrollableAddresses.MoveTo(dst.uncontrollableAddresses)
	gen.uncontrollableSlices.MoveTo(dst.uncontrollableSlices)
}
