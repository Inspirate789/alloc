package alloc

import (
	"github.com/Inspirate789/alloc/internal/generation"
)

var mainHypervisor *hypervisor

func init() { // TODO: init GC state and channel
	mainHypervisor = &hypervisor{
		mem: memory{
			movingGenerations: []*generation.Generation{
				generation.NewGeneration(), // young
				generation.NewGeneration(), // middle
				generation.NewGeneration(), // old
			},
			largeObjectGeneration: generation.NewGeneration(),
		},
	}

	go mainHypervisor.run()
}
