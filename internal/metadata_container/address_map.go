package metadata_container

import (
	"maps"
	"unsafe"
)

type address interface {
	Address() unsafe.Pointer
}

// AddressMap TODO: add RWMutex?
type AddressMap[V address] map[unsafe.Pointer]V

func NewAddressMap[V address]() AddressMap[V] {
	return make(AddressMap[V])
}

// TODO: lock on search

func (am AddressMap[V]) Add(value V) {
	am[value.Address()] = value
}

func (am AddressMap[V]) Search(addr unsafe.Pointer) (value V, exist bool) {
	value, exist = am[addr]
	return
}

func (am AddressMap[V]) Map(f func(value V)) {
	for _, val := range am {
		f(val)
	}
}

func (am AddressMap[V]) MoveTo(container any) {
	dst := container.(AddressMap[V]) // avoid?
	maps.Copy(dst, am)
	clear(am)
}

func (am AddressMap[V]) Delete(addr unsafe.Pointer) {
	delete(am, addr)
}
