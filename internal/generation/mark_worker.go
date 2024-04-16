package generation

import (
	"reflect"
	"slices"
	"unsafe"
)

type markWorker struct {
	gcID           uint64
	visited        map[unsafe.Pointer]bool
	searchMetadata func(addr unsafe.Pointer) (metadata *objectMetadata, exist bool)
}

func (mw markWorker) markObject(object *objectMetadata) (skip bool) {
	if object.finalized.Load() {
		return true
	}

	object.referenceCount++
	if object.lastMarkID == mw.gcID && mw.visited[object.address] {
		object.cycleReferenceSource = object
		return true
	} else {
		object.lastMarkID = mw.gcID
		mw.visited[object.address] = true
		return false
	}
}

func containerTypes() [6]reflect.Kind {
	return [6]reflect.Kind{
		reflect.Pointer, reflect.Interface, reflect.Array, reflect.Slice, reflect.Map, reflect.Struct,
	}
}

func (mw markWorker) extractNestedObjects(object reflect.Value) (nestedObjects []reflect.Value) {
	switch object.Kind() {
	case reflect.Pointer, reflect.Interface:
		nestedObjects = []reflect.Value{object.Elem()}
	case reflect.Array, reflect.Slice:
		containerTypes := containerTypes()
		if object.Len() != 0 && !slices.Contains(containerTypes[:], object.Index(0).Kind()) {
			break
		}
		for i := 0; i < object.Len(); i++ {
			nestedObjects = append(nestedObjects, object.Index(i))
		}
	case reflect.Map:
		containerTypes := containerTypes()
		if object.Len() != 0 && !slices.Contains(containerTypes[:], object.Index(0).Kind()) {
			break
		}
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
	object := reflect.NewAt(metadata.typeInfo, metadata.address).Elem()
	nestedObjects := mw.extractNestedObjects(object)
	for _, nestedObject := range nestedObjects {
		if !nestedObject.CanAddr() {
			continue
		}
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
	object.Lock()
	skip := mw.markObject(object)
	if !skip {
		nextObjects := mw.analyzeObject(object)
		object.Unlock()
		for _, nextObject := range nextObjects {
			if nextObject == object {
				continue
			}
			mw.processObject(nextObject)
			delete(mw.visited, nextObject.address)
			if nextObject.cycleReferenceSource == object {
				object.cycleReferenceCompleted = true
			} else if nextObject.cycleReferenceSource != nil && !nextObject.cycleReferenceCompleted {
				object.cycleReferenceSource = nextObject.cycleReferenceSource
			}
		}
	} else {
		object.Unlock()
	}

	return
}

func (mw markWorker) mark(objects <-chan *objectMetadata) {
	for object := range objects {
		object.referenceCount = -1
		mw.processObject(object)
	}
}
