package generation

import (
	"alloc"
	"reflect"
)

func extractNestedObjects(object reflect.Value) (nestedObjects []uintptr, nextObjects []reflect.Value) {
	switch object.Kind() {
	case reflect.Array, reflect.Slice:
		// if object.Len() != 0 && object.Index(0)
		for i := 0; i < object.Len(); i++ {
			// TODO
		}
	case reflect.Map:
		switch object.Type().Elem().Kind() {
		case reflect.Uintptr, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
			for i := 0; i < object.Len(); i++ {
				// TODO
			}
		}
	case reflect.Interface:
		value := object.Elem()
		nestedObjects = append(nestedObjects, value.UnsafeAddr())
		nextObjects = append(nextObjects, value)
	case reflect.Pointer, reflect.UnsafePointer:
		nextObjects = append(nextObjects, object.Elem())
	case reflect.Struct:
		// TODO
	}

	return
}

type color int

const (
	white = iota
	grey
	black
	red
)

type cycleDetector struct {
	coloredAddresses addressContainer[uintptr, color]
	curCycle         []uintptr
	cycles           [][]uintptr
}

func (d cycleDetector) detectCycles(object reflect.Value) {
	// TODO: colorize

	switch object.Kind() {
	case reflect.Pointer, reflect.Interface:
		d.detectCycles(object.Elem())
	case reflect.Array, reflect.Slice:
		for i := 0; i < object.Len(); i++ {
			d.detectCycles(object.Index(i))
		}
	case reflect.Map:
		mapIter := object.MapRange()
		for mapIter.Next() {
			d.detectCycles(mapIter.Key())
			d.detectCycles(mapIter.Value())
		}
	case reflect.Struct:
		for i := 0; i < object.NumField(); i++ {
			d.detectCycles(object.Field(i))
		}
	case reflect.UnsafePointer:
		ptr := object.UnsafePointer()
		p := object.Pointer()
		reflect.Indirect()
		// TODO
	case reflect.Uintptr:
		// TODO
	}

	// TODO: colorize
}

// roots must be pointers to the root objects
func detectCycleReferences(roots []any) [][]uintptr {
	detector := cycleDetector{
		coloredAddresses: alloc.addressMap[uintptr, color](make(map[uintptr]color, len(roots))),
	}

}
