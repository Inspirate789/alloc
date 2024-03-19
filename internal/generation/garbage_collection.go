package generation

import "unsafe"

func (gen *Generation) Contains(ptr unsafe.Pointer) bool {
	_, exist := gen.addresses.SearchByAddress(ptr)
	if !exist {
		_, exist = gen.slices.SearchByAddress(ptr)
	}
	return exist
}

func (gen *Generation) Cleanup(isObservableObject func(unsafe.Pointer) bool) {
	// TODO
}
