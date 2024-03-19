package generation

import (
	"runtime"
	"unsafe"
)

var gcMarkConcurrency = runtime.NumCPU()

func (gen *Generation) search(ptr unsafe.Pointer) bool {
	_, exist := gen.addresses.Search(ptr)
	if !exist {
		_, exist = gen.slices.Search(ptr)
	}
	return exist
}

func markObject(object *objectMetadata) (skip bool) {
	// TODO
	return true
}

func analyzeObject(object *objectMetadata) (nextObjects []*objectMetadata) {
	// TODO
	return nil
}

func mark(objects chan *objectMetadata) {
	for object := range objects {
		skip := markObject(object)
		if !skip {
			for _, nextObject := range analyzeObject(object) {
				objects <- nextObject
			}
		}
	}
}

func (gen *Generation) Cleanup() {
	objects := make(chan *objectMetadata)
	for i := 0; i < gcMarkConcurrency; i++ {
		go mark(objects)
	}

	gen.addresses.Map(func(metadata *objectMetadata) {
		objects <- metadata // TODO: movingMx? Lock?
	})

	gen.slices.Map(func(metadata *SliceMetadata) {
		objects <- &metadata.objectMetadata // TODO: movingMx? Lock?
	})

	close(objects)

	// TODO
}
