package generation

import (
	"alloc/limited_arena"
	"sync"
)

type addressContainer[K comparable, V any] interface {
	Set(key K, value V)
	Get(key K) (value V, exist bool)
	Delete(key K)
}

type Generation struct {
	arenas       []limited_arena.LimitedArena
	age          uint // the amount of garbage collection that a generation has survived
	collectionMx sync.Mutex
	addresses    addressContainer[uintptr, bool]
}

func AllocateObject[T any](gen *Generation) (get func() *T, finalize func()) {
	// choose the most loaded arena
	// allocate object
	return nil, nil // TODO
}

func AllocateSlice[T any](gen *Generation, len, cap int) (get func() []T, finalize func()) {
	// choose the most loaded arena
	// allocate slice
	return nil, nil // TODO
}

func (gen *Generation) Contains(objectPtr uintptr) bool {
	_, exist := gen.addresses.Get(objectPtr)
	return exist
}

func (gen *Generation) Cleanup(isObservableObject func(uintptr) bool) {
	// TODO
}
