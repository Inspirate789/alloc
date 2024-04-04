package alloc

import (
	"context"
	"github.com/Inspirate789/alloc/internal/generation"
	"math/rand/v2"
	"sync"
)

type gcState struct {
	lastCollectedGenerations int
}

type hypervisor struct {
	gcLock sync.Mutex
	gcCtx  context.Context
	mem    memory
	gcState
}

// TODO: choose when gc is needed

func (h *hypervisor) run() {

}

func mergeSearchFunctions[F ~func(K) (V, bool), K comparable, V any](functions []F) F {
	return func(key K) (value V, exist bool) {
		for _, search := range functions {
			value, exist = search(key)
			if exist {
				return
			}
		}
		return
	}
}

func (h *hypervisor) gc() {
	generationsCount := (h.lastCollectedGenerations+1)%(len(h.mem.movingGenerations)+1) + 1
	h.lastCollectedGenerations = generationsCount
	generations := append(h.mem.movingGenerations, h.mem.largeObjectGeneration)

	searchFunctions := make([]generation.SearchFunc, 0, generationsCount)
	for i := 0; i < generationsCount; i++ {
		searchFunctions = append(searchFunctions, generations[i].SearchObject)
	}
	searchFunc := mergeSearchFunctions(searchFunctions)

	gcID := rand.Uint64()
	wg := sync.WaitGroup{}
	for i := 0; i < generationsCount; i++ {
		i := i // TODO: remove
		wg.Add(1)
		go func() {
			generations[i].Mark(gcID, searchFunc)
			generations[i].Compact()
			wg.Done()
		}()
	}
	wg.Wait()
}

func GC() {
	if mainHypervisor.gcLock.TryLock() {
		ctx, cancel := context.WithCancel(context.Background())
		mainHypervisor.gcCtx = ctx
		mainHypervisor.gc()
		mainHypervisor.gcLock.Unlock()
		cancel()
	} else {
		<-mainHypervisor.gcCtx.Done()
	}
}
