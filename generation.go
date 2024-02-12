package alloc

import (
	"alloc/limited_arena"
	"sync"
)

type generation struct {
	arenas       []limited_arena.LimitedArena
	age          uint // the amount of garbage collection that a generation has survived
	collectionMx sync.Mutex
	// TODO: addresses container
}

func allocateGenerationObject[T any](gen *generation) (func() *T, func()) {
	return nil, nil // TODO
}

func allocateGenerationSlice[T any](gen *generation) (func() []T, func()) {
	return nil, nil // TODO
}
