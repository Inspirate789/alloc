package generation

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestMarkOneSimpleObject(t *testing.T) {
	t.Parallel()
	// arrange
	object := 7
	metadata := &ObjectMetadata{
		address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
	}
	objects := map[unsafe.Pointer]*ObjectMetadata{
		metadata.address: metadata,
	}

	input := make(chan *ObjectMetadata, 1)
	for _, object := range objects {
		input <- object
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]bool),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
		gcMetadata: gcMetadata{
			lastMarkID:           gcID,
			cyclicallyReferenced: false,
			referenceCount:       0,
		},
	}
	if !reflect.DeepEqual(metadata, expectedMetadata) {
		t.Errorf("metadata = %+v, want %+v", metadata, expectedMetadata)
	}
}

type testObject struct {
	id        int
	nilPtr    *int
	nonNilPtr *int
	fieldPtr  **int
	self      *testObject
	other     *testObject
}

func TestMarkOneStruct(t *testing.T) {
	t.Parallel()
	// arrange
	object := testObject{id: 1}
	object.nonNilPtr = &object.id
	object.fieldPtr = &object.nonNilPtr
	object.self = &object
	metadata := &ObjectMetadata{
		address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
	}
	objects := map[unsafe.Pointer]*ObjectMetadata{
		metadata.address: metadata,
	}

	input := make(chan *ObjectMetadata, 1)
	for _, object := range objects {
		input <- object
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]bool),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
		gcMetadata: gcMetadata{
			lastMarkID:           gcID,
			cyclicallyReferenced: false,
			referenceCount:       0,
		},
	}
	if !reflect.DeepEqual(metadata, expectedMetadata) {
		t.Errorf("metadata = %+v, want %+v", metadata, expectedMetadata)
	}
}

func TestMarkStructSlice(t *testing.T) {
	t.Parallel()
	// arrange
	const objectsCount = 10
	elements := make([]testObject, objectsCount)
	objects := make(map[unsafe.Pointer]*ObjectMetadata)
	for i := range objectsCount {
		elements[i] = testObject{
			id: i,
		}
		elements[i].nonNilPtr = &elements[i].id
		elements[i].fieldPtr = &elements[i].nonNilPtr
		elements[i].self = &elements[i]
		if i != 0 {
			elements[i].other = &elements[i-1]
		}
		elements = append(elements, elements[i])
		metadata := &ObjectMetadata{
			address:  unsafe.Pointer(&elements[i]),
			typeInfo: reflect.TypeOf(elements[i]),
		}
		objects[metadata.address] = metadata
	}

	rootMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
	}
	objects[rootMetadata.address] = rootMetadata

	input := make(chan *ObjectMetadata, objectsCount+1)
	for _, metadata := range objects {
		input <- metadata
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]bool),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
		gcMetadata: gcMetadata{
			lastMarkID:           gcID,
			cyclicallyReferenced: false,
			referenceCount:       0,
		},
	}
	if !reflect.DeepEqual(rootMetadata, expectedMetadata) {
		t.Errorf("metadata = %+v, want %+v", rootMetadata, expectedMetadata)
	}
}

func TestMarkStructCycledSlice(t *testing.T) {
	t.Parallel()
	// arrange
	const objectsCount = 2
	elements := make([]testObject, objectsCount)
	objects := make(map[unsafe.Pointer]*ObjectMetadata)
	for i := range objectsCount {
		elements[i] = testObject{
			id: i,
		}
		elements[i].nonNilPtr = &elements[i].id
		elements[i].fieldPtr = &elements[i].nonNilPtr
		elements[i].self = &elements[i]
		elements[i].other = &elements[(i+1)%objectsCount]
		metadata := &ObjectMetadata{
			address:  unsafe.Pointer(&elements[i]),
			typeInfo: reflect.TypeOf(elements[i]),
		}
		objects[metadata.address] = metadata
	}

	rootMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
	}
	objects[rootMetadata.address] = rootMetadata

	input := make(chan *ObjectMetadata, objectsCount+1)
	for _, metadata := range objects {
		input <- metadata
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]bool),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
		gcMetadata: gcMetadata{
			lastMarkID:           gcID,
			cyclicallyReferenced: false,
			referenceCount:       0,
		},
	}
	if !reflect.DeepEqual(rootMetadata, expectedMetadata) {
		t.Errorf("metadata = %+v, want %+v", rootMetadata, expectedMetadata)
	}
	delete(objects, rootMetadata.address)
	for _, metadata := range objects {
		if metadata.lastMarkID != expectedMetadata.lastMarkID {
			t.Errorf("metadata.lastMarkID = %+v, want %+v", metadata.lastMarkID, expectedMetadata.lastMarkID)
		}
		if !metadata.cyclicallyReferenced {
			t.Errorf("metadata.cyclicallyReferenced = false, want true")
		}
		if metadata.referenceCount != 2 {
			t.Errorf("metadata.referenceCount = %+v, want 2", metadata.referenceCount)
		}
	}
}
