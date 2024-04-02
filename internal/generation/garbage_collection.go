package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"runtime"
	"slices"
	"sync"
	"unsafe"
)

var gcMarkConcurrency = (runtime.NumCPU() + 1) / 2

func (gen *Generation) searchMetadata(addr unsafe.Pointer) (metadata *objectMetadata, exist bool) {
	metadata, exist = gen.addresses.Search(addr)
	if !exist {
		var sliceMetadata *SliceMetadata
		sliceMetadata, exist = gen.slices.Search(addr)
		if sliceMetadata != nil {
			metadata = &sliceMetadata.objectMetadata
		}
	}
	if !exist {
		metadata, exist = gen.uncontrollableAddresses.Search(addr)
	}
	if !exist {
		var sliceMetadata *SliceMetadata
		sliceMetadata, exist = gen.uncontrollableSlices.Search(addr)
		if sliceMetadata != nil {
			metadata = &sliceMetadata.objectMetadata
		}
	}
	return
}

type SearchFunc func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool)

func (gen *Generation) SearchObjectFunc() SearchFunc {
	return gen.searchMetadata
}

func (gen *Generation) Mark(gcID uint64, searchMetadata SearchFunc) {
	if searchMetadata == nil {
		searchMetadata = gen.searchMetadata
	} else {
		searchMetadata = func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool) {
			metadata, exist = gen.searchMetadata(addr)
			if !exist {
				metadata, exist = searchMetadata(addr)
			}
			return
		}
	}

	objects := make(chan *objectMetadata)
	wg := sync.WaitGroup{}
	for i := 0; i < gcMarkConcurrency; i++ {
		mw := markWorker{
			gcID:           gcID,
			visited:        make(map[unsafe.Pointer]bool),
			searchMetadata: searchMetadata,
		}
		wg.Add(1)
		go func() {
			mw.mark(objects)
			wg.Done()
		}()
	}

	gen.addresses.Map(func(metadata *objectMetadata) {
		objects <- metadata
	})

	gen.slices.Map(func(metadata *SliceMetadata) {
		objects <- &metadata.objectMetadata
	})

	close(objects)
	wg.Wait()
}

// (cyclicallyReferenced && referenceCount == 0) || finalized ==> dead object
func isGarbage(object *objectMetadata) bool {
	return (object.cyclicallyReferenced && object.referenceCount == 0) || object.finalized.Load() // TODO: use && instead of || ?
}

func (gen *Generation) detectGarbageArenas() []*limited_arena.Arena {
	arenaObjectsCount := make(map[*limited_arena.Arena]int, len(gen.arenas))
	garbageObjectsCount := make(map[*limited_arena.Arena]int, len(gen.arenas))

	gen.addresses.Map(func(object *objectMetadata) {
		arenaObjectsCount[object.arena]++
		if isGarbage(object) {
			garbageObjectsCount[object.arena]++
		}
	})

	gen.slices.Map(func(object *SliceMetadata) {
		arenaObjectsCount[object.arena]++
		if isGarbage(&object.objectMetadata) {
			garbageObjectsCount[object.arena]++
		}
	})

	garbageArenas := make([]*limited_arena.Arena, 0)
	for arena, count := range arenaObjectsCount {
		if garbageObjectsCount[arena] == count && count != 0 {
			garbageArenas = append(garbageArenas, arena)
		}
	}

	return garbageArenas
}

func (gen *Generation) Compact() bool {
	garbageArenas := gen.detectGarbageArenas()
	if len(garbageArenas) == 0 {
		return false
	}

	for offset, arena := range garbageArenas { // TODO: lock generation?
		index := slices.Index(gen.arenas, *arena)
		if index == -1 {
			panic("unknown arena") // TODO: remove
		}
		tailIndex := len(gen.arenas) - offset - 1
		gen.arenas[index] = gen.arenas[tailIndex]
		gen.arenas[tailIndex].Free()
		gen.arenas[tailIndex] = limited_arena.Arena{}
	}

	gen.arenas = gen.arenas[:len(gen.arenas)-len(garbageArenas)]

	return true
}
