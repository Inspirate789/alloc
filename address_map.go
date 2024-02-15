package alloc

type addressMap[K comparable, V any] map[K]V // TODO: whether map is the best choice in this case?

func (am addressMap[K, V]) Set(key K, value V) {
	am[key] = value
}

func (am addressMap[K, V]) Get(key K) (value V, exist bool) {
	value, exist = am[key]
	return
}

func (am addressMap[K, V]) Delete(key K) {
	delete(am, key)
}
