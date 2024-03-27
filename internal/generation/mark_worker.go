package generation

import (
	"reflect"
	"unsafe"
)

type markWorker struct {
	id             int
	gcID           uint64
	visited        map[unsafe.Pointer]bool
	searchMetadata func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool)
}

func (mw markWorker) markObject(object *objectMetadata) (skip bool) {
	if object.lastMarkID == mw.gcID && mw.visited[object.Addr] {
		object.cyclicallyReferenced = true
		return true
	} else {
		object.lastMarkID = mw.gcID
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
