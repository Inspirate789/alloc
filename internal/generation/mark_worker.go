package generation

import (
	"reflect"
	"slices"
	"unsafe"
)

type markWorker struct {
	gcID           uint64
	visited        map[unsafe.Pointer]bool
	searchMetadata SearchFunc
}

func (mw markWorker) markObject(object *ObjectMetadata, incRefCount bool) (rcSrc *ObjectMetadata, skip, visited bool) {
	if object.finalized.Load() {
		return nil, true, false
	}

	if incRefCount {
		object.referenceCount++
	}

	if object.lastMarkID == mw.gcID {
		if mw.visited[object.address] {
			object.cyclicallyReferenced = true
			return object, true, true
		} else {
			mw.visited[object.address] = true
			return nil, false, true
		}
	} else {
		object.lastMarkID = mw.gcID
		mw.visited[object.address] = true
		return nil, false, false
	}
}

func containerTypes() [6]reflect.Kind {
	return [6]reflect.Kind{
		reflect.Pointer, reflect.Interface, reflect.Array, reflect.Slice, reflect.Map, reflect.Struct,
	}
}

func extractNestedObjects(object reflect.Value) (nestedObjects []reflect.Value) {
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

func (mw markWorker) extractMetadata(object reflect.Value) (metadata *ObjectMetadata, exist bool) {
	kind := object.Kind()

	if kind == reflect.Interface || kind == reflect.Pointer { // object points to known object
		elem := object.Elem()
		if elem.CanAddr() {
			ptr := elem.Addr().UnsafePointer()
			if metadata, exist = mw.searchMetadata(ptr); exist {
				return
			}
		}
	}

	if object.CanAddr() { // object is known
		ptr := object.Addr().UnsafePointer()
		if metadata, exist = mw.searchMetadata(ptr); exist {
			return
		}
	}

	if kind == reflect.Interface { // object is getter
		field := object.Elem().FieldByName("metadata").Interface()
		if metadata, exist = field.(*ObjectMetadata); exist {
			return
		}
	}

	return
}

func (mw markWorker) analyzeObject(metadata *ObjectMetadata) (nextObjects []*ObjectMetadata) {
	object := reflect.NewAt(metadata.typeInfo, metadata.address).Elem()
	nestedObjects := extractNestedObjects(object)
	for _, nestedObject := range nestedObjects {
		if nestedMetadata, exist := mw.extractMetadata(nestedObject); exist {
			nextObjects = append(nextObjects, nestedMetadata)
		} else {
			// I can't trace it ¯\_(ツ)_/¯
		}
	}

	return
}

func (mw markWorker) processObject(object *ObjectMetadata, visited bool) (rcSrcs []*ObjectMetadata) {
	object.Lock()
	rcSrc, skip, visited := mw.markObject(object, !visited)
	if !skip {
		nextObjects := mw.analyzeObject(object)
		object.Unlock()
		for _, nextObject := range nextObjects {
			if nextObject == object {
				continue
			}
			rcSrcs = append(rcSrcs, mw.processObject(nextObject, visited)...)
			delete(mw.visited, nextObject.address)
			if len(rcSrcs) != 0 {
				object.cyclicallyReferenced = true
				rcSrcs = slices.DeleteFunc(rcSrcs, func(metadata *ObjectMetadata) bool {
					return metadata == object
				})
			}
		}
	} else {
		object.Unlock()
		rcSrcs = append(rcSrcs, rcSrc)
	}

	return
}

func (mw markWorker) mark(objects <-chan *ObjectMetadata) {
	for object := range objects {
		mw.processObject(object, false)
		delete(mw.visited, object.address)
	}
}
