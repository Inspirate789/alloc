package generation

import (
	"alloc/limited_arena"
	"sync"
)

type Generation struct {
	arenas       []limited_arena.LimitedArena
	age          uint // the amount of garbage collection that a generation has survived
	collectionMx sync.Mutex
	// TODO: addresses container
}
