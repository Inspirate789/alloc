package alloc

import (
	"context"
	"github.com/Inspirate789/alloc/internal/generation"
)

const (
	arenaSignalsQueueLen = 10
	startHeapSize        = 4
	startHeapTarget      = 15
)

var mainHypervisor *hypervisor

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	arenaSignals := make(chan struct{}, arenaSignalsQueueLen)
	mainHypervisor = &hypervisor{
		gcCtx:        ctx,
		gcCancel:     cancel,
		arenaSignals: arenaSignals,
		gcState: gcState{
			lastCollectedGenerations: -1,
			curHeapSize:              startHeapSize,
			heapTarget:               startHeapTarget,
		},
		mem: memory{
			movingGenerations: []*generation.Generation{
				generation.NewGeneration(), // young
				generation.NewGeneration(), // middle
				generation.NewGeneration(), // old
			},
			largeObjectGeneration: generation.NewGeneration(),
		},
	}
	Debugger.arenasAllocated.Store(3)

	go mainHypervisor.run()
}
