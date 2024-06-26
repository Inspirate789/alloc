package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"runtime"
	"slices"
	"sync"
	"unsafe"
)

var gcMarkConcurrency = (runtime.NumCPU() + 1) / 2

func (gen *Generation) SearchObject(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
	metadata, exist = gen.addresses.Search(addr)
	if !exist {
		var sliceMetadata *SliceMetadata
		sliceMetadata, exist = gen.slices.Search(addr)
		if sliceMetadata != nil {
			metadata = &sliceMetadata.ObjectMetadata
		}
	}
	if !exist {
		metadata, exist = gen.uncontrollableAddresses.Search(addr)
	}
	if !exist {
		var sliceMetadata *SliceMetadata
		sliceMetadata, exist = gen.uncontrollableSlices.Search(addr)
		if sliceMetadata != nil {
			metadata = &sliceMetadata.ObjectMetadata
		}
	}
	return
}

type SearchFunc func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool)

func (gen *Generation) Mark(gcID uint64, searchMetadata SearchFunc) {
	search := func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
		metadata, exist = gen.SearchObject(addr)
		if !exist {
			metadata, exist = searchMetadata(addr)
		}
		return
	}

	objects := make(chan *ObjectMetadata)
	wg := sync.WaitGroup{}
	for range gcMarkConcurrency {
		mw := markWorker{
			gcID:           gcID,
			visited:        make(map[unsafe.Pointer]struct{}),
			searchMetadata: search,
		}
		wg.Add(1)
		go func() {
			mw.mark(objects)
			wg.Done()
		}()
	}

	gen.addresses.Map(func(metadata *ObjectMetadata) {
		metadata.cyclicallyReferenced = false
		metadata.referenceCount--
		objects <- metadata
	})

	gen.slices.Map(func(metadata *SliceMetadata) {
		metadata.cyclicallyReferenced = false
		metadata.referenceCount--
		objects <- &metadata.ObjectMetadata
	})

	close(objects)
	wg.Wait()
}

// (cyclicallyReferenced && referenceCount <= 1) || finalized ==> dead object
func isGarbage(object *ObjectMetadata) bool {
	return (object.cyclicallyReferenced && object.referenceCount <= 1) || object.Finalized.Load()
}

func (gen *Generation) cleanUncontrollableObjects() {
	garbageObjects := make([]unsafe.Pointer, 0)
	gen.uncontrollableAddresses.Map(func(object *ObjectMetadata) {
		if object.Finalized.Load() {
			garbageObjects = append(garbageObjects, object.Address)
		}
	})
	gen.uncontrollableAddresses.Delete(garbageObjects)

	garbageSlices := make([]unsafe.Pointer, 0)
	gen.uncontrollableSlices.Map(func(slice *SliceMetadata) {
		if slice.Finalized.Load() {
			garbageSlices = append(garbageSlices, slice.Address)
		}
	})
	gen.uncontrollableSlices.Delete(garbageSlices)
}

func (gen *Generation) detectGarbageArenas() []*limited_arena.Arena {
	arenaObjectsCount := make(map[*limited_arena.Arena]int, len(gen.arenas))
	garbageObjectsCount := make(map[*limited_arena.Arena]int, len(gen.arenas))

	garbageAddresses := make([]unsafe.Pointer, 0)
	garbageSlices := make([]unsafe.Pointer, 0)

	gen.addresses.Map(func(object *ObjectMetadata) {
		arenaObjectsCount[object.arena]++
		if isGarbage(object) {
			garbageObjectsCount[object.arena]++
			garbageAddresses = append(garbageAddresses, object.Address)
		}
	})

	gen.slices.Map(func(object *SliceMetadata) {
		arenaObjectsCount[object.arena]++
		if isGarbage(&object.ObjectMetadata) {
			garbageObjectsCount[object.arena]++
			garbageSlices = append(garbageSlices, object.Address)
		}
	})

	garbageArenas := make([]*limited_arena.Arena, 0)
	for arena, count := range arenaObjectsCount {
		if garbageObjectsCount[arena] == count && count != 0 {
			garbageArenas = append(garbageArenas, arena)
		}
	}

	// go func() {
	gen.addresses.Delete(garbageAddresses)
	gen.slices.Delete(garbageSlices)
	gen.cleanUncontrollableObjects()
	// }()

	return garbageArenas
}

func (gen *Generation) Sweep() (int, int) {
	gen.arenasMx.Lock()

	sizeBefore := len(gen.arenas)

	garbageArenas := gen.detectGarbageArenas()
	if len(garbageArenas) == 0 {
		return sizeBefore, sizeBefore
	}

	for offset, arena := range garbageArenas {
		index := slices.Index(gen.arenas, *arena)
		//if index == -1 {
		//	panic("unknown arena")
		//}
		tailIndex := len(gen.arenas) - offset - 1
		gen.arenas[index] = gen.arenas[tailIndex]
		gen.arenas[tailIndex].Free()
		gen.arenas[tailIndex] = limited_arena.Arena{}
	}

	gen.arenas = gen.arenas[:len(gen.arenas)-len(garbageArenas)]

	gen.arenasMx.Unlock()

	return sizeBefore, sizeBefore - len(garbageArenas)
}
