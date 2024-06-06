package metadata_container

import (
	"maps"
	"sync"
	"unsafe"
)

const startMapSize = 5000 // TODO: remove?

type addressMap[V any] map[unsafe.Pointer]V

type AddressContainer[V any] struct {
	addressMap[V]
	lock *sync.RWMutex
}

func NewAddressContainer[V any]() AddressContainer[V] {
	return AddressContainer[V]{
		addressMap: make(addressMap[V], startMapSize),
		lock:       new(sync.RWMutex),
	}
}

func (ac AddressContainer[V]) Add(key unsafe.Pointer, value V) {
	// ac.lock.Lock()
	ac.addressMap[key] = value
	// ac.lock.Unlock()
}

func (ac AddressContainer[V]) Search(addr unsafe.Pointer) (value V, exist bool) {
	ac.lock.RLock()
	value, exist = ac.addressMap[addr]
	ac.lock.RUnlock()
	return
}

func (ac AddressContainer[V]) Map(f func(value V)) {
	ac.lock.RLock()
	for _, val := range ac.addressMap {
		f(val)
	}
	ac.lock.RUnlock()
}

func (ac AddressContainer[V]) MoveTo(container any) {
	dst := container.(AddressContainer[V]) // avoid?

	ac.lock.Lock()
	dst.lock.Lock()

	maps.Copy(dst.addressMap, ac.addressMap)
	clear(ac.addressMap)

	dst.lock.Unlock()
	ac.lock.Unlock()
}

func (ac AddressContainer[V]) Move(old, new unsafe.Pointer) {
	ac.lock.Lock()
	value, exist := ac.addressMap[old]
	if exist {
		delete(ac.addressMap, old)
		ac.addressMap[new] = value
	}
	ac.lock.Unlock()
}

func (ac AddressContainer[V]) Delete(addresses []unsafe.Pointer) {
	ac.lock.Lock()
	for _, addr := range addresses {
		delete(ac.addressMap, addr)
	}
	ac.lock.Unlock()
}
