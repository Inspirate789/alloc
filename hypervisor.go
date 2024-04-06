package alloc

import (
	"context"
	"github.com/Inspirate789/alloc/internal/generation"
	"math"
	"math/rand/v2"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type gcState struct {
	lastCollectedGenerations int
	curHeapSize              int
	heapTarget               int
}

type hypervisor struct {
	gcLock       sync.Mutex
	gcCtx        context.Context
	gcCancel     context.CancelFunc
	mem          memory
	arenaSignals <-chan struct{}
	gcState
}

func (h *hypervisor) run() {
	for range h.arenaSignals {
		if h.gcLock.TryLock() {
			h.curHeapSize++
			curHeapSize := h.curHeapSize
			heapTarget := h.heapTarget
			h.gcLock.Unlock()
			if curHeapSize >= heapTarget {
				h.gc()
			}
		}
		Debugger.arenasAllocated.Add(1)
	}
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

func (h *hypervisor) mergeGenerations(sizesBefore, sizesAfter []int) { // TODO: fill debugger.arenasFreed
	// TODO
}

func gogc() int {
	percent := debug.SetGCPercent(100)
	debug.SetGCPercent(percent)

	return percent
}

func (h *hypervisor) calculateNewHeapTarget() {
	sizes := make([]int, len(h.mem.movingGenerations)+1)

	for i := range h.mem.movingGenerations {
		sizes[i] = h.mem.movingGenerations[i].Size()
	}
	sizes[len(sizes)-1] = h.mem.largeObjectGeneration.Size()

	sizeTotal := 0
	for _, size := range sizes {
		sizeTotal += size
	}

	h.curHeapSize = sizeTotal
	h.heapTarget = int(math.Round(float64(h.curHeapSize) * (1 + float64(gogc())/100)))
}

func (h *hypervisor) gc() {
	runtime.GC()

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
	sizesBefore := make([]int, len(generations))
	sizesAfter := make([]int, len(generations))
	for i := 0; i < generationsCount; i++ {
		i := i // TODO: remove
		wg.Add(1)
		go func() {
			generations[i].Mark(gcID, searchFunc)
			sizesBefore[i], sizesAfter[i] = generations[i].Compact()
			wg.Done()
		}()
	}

	for i := generationsCount; i < len(generations); i++ {
		size := generations[i].Size()
		sizesBefore[i] = size
		sizesAfter[i] = size
	}

	wg.Wait()

	h.mergeGenerations(sizesBefore, sizesAfter)
	h.calculateNewHeapTarget()
}

// GC runs a garbage collection and blocks the caller until the
// garbage collection is complete. It may also block the entire
// program.
func GC() {
	if mainHypervisor.gcLock.TryLock() {
		mainHypervisor.gc()

		ctx, cancel := context.WithCancel(context.Background())
		mainHypervisor.gcCancel()
		mainHypervisor.gcCancel = cancel

		Debugger.numCC.Add(1)
		Debugger.lastCC.Store(time.Now().Unix())

		mainHypervisor.gcLock.Unlock()

		mainHypervisor.gcCtx = ctx
	} else {
		<-mainHypervisor.gcCtx.Done()
	}
}
