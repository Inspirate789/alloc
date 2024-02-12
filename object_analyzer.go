package alloc

import "reflect"

func extractObjectPointers(object any) []uintptr {
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			// TODO
		}
	case reflect.Interface:
		// TODO
	case reflect.Map:
		// TODO
	case reflect.Pointer, reflect.UnsafePointer:
		// TODO
	case reflect.Struct:
		// TODO
	}

	return nil
}
