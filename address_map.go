package alloc

import "unsafe"

type address interface {
	Address() unsafe.Pointer
}

type addressMap[K comparable, V address] map[K]V // TODO: whether map is the best choice in this case?

func (am addressMap[K, V]) Set(key K, value V) {
	am[key] = value
}

func (am addressMap[K, V]) Get(key K) (value V, exist bool) {
	value, exist = am[key]
	return
}

func (am addressMap[K, V]) SearchByAddress(addr unsafe.Pointer) (value V, exist bool) {
	for _, v := range am {
		if v.Address() == addr {
			return v, true
		}
	}
	return
}

func (am addressMap[K, V]) Delete(key K) {
	delete(am, key)
}
