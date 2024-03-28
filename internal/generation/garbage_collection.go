package generation

import (
	"runtime"
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
			id:             i,
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

// TODO: move and compact
