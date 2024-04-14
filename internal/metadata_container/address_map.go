package metadata_container

import (
	"maps"
	"sync"
	"unsafe"
)

type address interface {
	Address() unsafe.Pointer
}

type addressMap[V address] map[unsafe.Pointer]V

type AddressContainer[V address] struct {
	addressMap[V]
	lock *sync.RWMutex
}

func NewAddressContainer[V address]() AddressContainer[V] {
	return AddressContainer[V]{
		addressMap: make(addressMap[V]),
		lock:       new(sync.RWMutex),
	}
}

func (ac AddressContainer[V]) Add(value V) {
	ac.lock.Lock()
	ac.addressMap[value.Address()] = value
	ac.lock.Unlock()
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
	ac.lock.Lock()
	dst := container.(addressMap[V]) // avoid?
	maps.Copy(dst, ac.addressMap)
	clear(ac.addressMap)
	ac.lock.Unlock()
}

func (ac AddressContainer[V]) Delete(addresses []unsafe.Pointer) {
	ac.lock.Lock()
	for _, addr := range addresses {
		delete(ac.addressMap, addr)
	}
	ac.lock.Unlock()
}
