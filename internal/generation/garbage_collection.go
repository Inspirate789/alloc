package generation

import (
	"math/rand"
	"reflect"
	"runtime"
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
	return
}

func (gen *Generation) SearchObjectFunc() func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool) {
	return gen.searchMetadata
}

type markWorker struct {
	id             int
	gcID           uint64
	visited        map[unsafe.Pointer]bool
	searchMetadata func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool)
}

func (mw markWorker) markObject(object *objectMetadata) (skip bool) {
	if object.lastGcID == mw.gcID && mw.visited[object.Addr] {
		object.cyclicallyReferenced = true
		return true
	} else {
		object.lastGcID = mw.gcID
		mw.visited[object.Addr] = true
		return false
	}
}

func (mw markWorker) extractNestedObjects(object reflect.Value) (nestedObjects []reflect.Value) {
	switch object.Kind() {
	case reflect.Pointer, reflect.Interface:
		nestedObjects = []reflect.Value{object.Elem()}
	case reflect.Array, reflect.Slice:
		for i := 0; i < object.Len(); i++ {
			nestedObjects = append(nestedObjects, object.Index(i))
		}
	case reflect.Map:
		mapIter := object.MapRange()
		for mapIter.Next() {
			nestedObjects = append(nestedObjects, mapIter.Key(), mapIter.Value())
		}
	case reflect.Struct:
		for i := 0; i < object.NumField(); i++ {
			nestedObjects = append(nestedObjects, object.Field(i))
		}
	}
	return
}

func (mw markWorker) analyzeObject(metadata *objectMetadata) (nextObjects []*objectMetadata) {
	object := reflect.NewAt(metadata.typeInfo, metadata.Addr).Elem()
	nestedObjects := mw.extractNestedObjects(object)
	for _, nestedObject := range nestedObjects {
		addr := nestedObject.Addr().UnsafePointer()
		if nestedMetadata, exist := mw.searchMetadata(addr); exist {
			nextObjects = append(nextObjects, nestedMetadata)
		} else {
			// I can't trace it ¯\_(ツ)_/¯
		}
	}

	return
}

func (mw markWorker) processObject(object *objectMetadata) {
	skip := mw.markObject(object)
	if !skip {
		for _, nextObject := range mw.analyzeObject(object) {
			mw.processObject(nextObject)
		}
	}

	return
}

func (mw markWorker) mark(objects <-chan *objectMetadata) {
	for object := range objects {
		mw.processObject(object)
	}
}

func (gen *Generation) Mark(searchMetadata func(unsafe.Pointer) (*objectMetadata, bool)) {
	gcID := rand.Uint64()

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
	for i := 0; i < gcMarkConcurrency; i++ {
		mw := markWorker{
			id:             i,
			gcID:           gcID,
			visited:        make(map[unsafe.Pointer]bool),
			searchMetadata: searchMetadata,
		}
		go mw.mark(objects)
	}

	gen.addresses.Map(func(metadata *objectMetadata) {
		objects <- metadata // TODO: movingMx? Lock?
	})

	gen.slices.Map(func(metadata *SliceMetadata) {
		objects <- &metadata.objectMetadata // TODO: movingMx? Lock?
	})

	close(objects)
}

// TODO: compact
