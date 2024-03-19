package metadata_container

import "unsafe"

type address interface {
	Address() unsafe.Pointer
}

type MetadataMap[K comparable, V address] map[K]V // TODO: whether map is the best choice in this case?

func NewMetadataMap[K comparable, V address](cap int) MetadataMap[K, V] {
	return make(MetadataMap[K, V], cap)
}

func (mm MetadataMap[K, V]) Set(key K, value V) {
	mm[key] = value
}

func (mm MetadataMap[K, V]) Get(key K) (value V, exist bool) {
	value, exist = mm[key]
	return
}

func (mm MetadataMap[K, V]) SearchByAddress(addr unsafe.Pointer) (value V, exist bool) {
	for _, v := range mm {
		if v.Address() == addr {
			return v, true
		}
	}
	return
}

func (mm MetadataMap[K, V]) Delete(key K) {
	delete(mm, key)
}
