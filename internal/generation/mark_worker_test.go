package generation

import (
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestObjectMetadata(t *testing.T) {
	t.Log(unsafe.Sizeof(sync.RWMutex{}), reflect.TypeOf(sync.RWMutex{}).Size())
	t.Log(unsafe.Sizeof(unsafe.Pointer(nil)), reflect.TypeOf(unsafe.Pointer(nil)).Size())
	t.Log(unsafe.Sizeof(reflect.TypeOf(1)), reflect.TypeOf(reflect.TypeOf(1)).Size())
	t.Log(unsafe.Sizeof(&limited_arena.Arena{}), reflect.TypeOf(&limited_arena.Arena{}).Size())
	t.Log(unsafe.Sizeof(uint64(0)), reflect.TypeOf(uint64(0)).Size())
	t.Log(unsafe.Sizeof(false), reflect.TypeOf(false).Size())
	t.Log(unsafe.Sizeof(int(0)), reflect.TypeOf(int(0)).Size())
	t.Log(unsafe.Sizeof(atomic.Bool{}), reflect.TypeOf(atomic.Bool{}).Size())
	t.Log(unsafe.Sizeof(gcMetadata{}), reflect.TypeOf(gcMetadata{}).Size())
	t.Log(unsafe.Sizeof(ObjectMetadata{}), reflect.TypeOf(ObjectMetadata{}).Size())
}

func TestMarkOneSimpleObject(t *testing.T) {
	t.Parallel()
	// arrange
	object := 7
	metadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
	}
	objects := map[unsafe.Pointer]*ObjectMetadata{
		metadata.Address: metadata,
	}

	input := make(chan *ObjectMetadata, 1)
	for _, object := range objects {
		object.cyclicallyReferenced = false
		object.referenceCount--
		input <- object
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]struct{}),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&object),
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
		Address:  unsafe.Pointer(&object),
		typeInfo: reflect.TypeOf(object),
	}
	objects := map[unsafe.Pointer]*ObjectMetadata{
		metadata.Address: metadata,
	}

	input := make(chan *ObjectMetadata, 1)
	for _, object := range objects {
		object.cyclicallyReferenced = false
		object.referenceCount--
		input <- object
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]struct{}),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&object),
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
			Address:  unsafe.Pointer(&elements[i]),
			typeInfo: reflect.TypeOf(elements[i]),
		}
		objects[metadata.Address] = metadata
	}

	rootMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
	}
	objects[rootMetadata.Address] = rootMetadata

	input := make(chan *ObjectMetadata, objectsCount+1)
	for _, metadata := range objects {
		metadata.cyclicallyReferenced = false
		metadata.referenceCount--
		input <- metadata
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]struct{}),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&elements),
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
			Address:  unsafe.Pointer(&elements[i]),
			typeInfo: reflect.TypeOf(elements[i]),
		}
		objects[metadata.Address] = metadata
	}

	rootMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&elements),
		typeInfo: reflect.TypeOf(elements),
	}
	objects[rootMetadata.Address] = rootMetadata

	input := make(chan *ObjectMetadata, objectsCount+1)
	for _, metadata := range objects {
		metadata.cyclicallyReferenced = false
		metadata.referenceCount--
		input <- metadata
	}
	close(input)

	const gcID = 1
	mw := markWorker{
		gcID:    gcID,
		visited: make(map[unsafe.Pointer]struct{}),
		searchMetadata: func(addr unsafe.Pointer) (metadata *ObjectMetadata, exist bool) {
			metadata, exist = objects[addr]
			return
		},
	}

	// act
	mw.mark(input)

	// assert
	expectedMetadata := &ObjectMetadata{
		Address:  unsafe.Pointer(&elements),
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
	delete(objects, rootMetadata.Address)
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
