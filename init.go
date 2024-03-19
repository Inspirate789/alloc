package alloc

import "github.com/Inspirate789/alloc/generation"

var mainHypervisor hypervisor

func init() {
	mainHypervisor = hypervisor{
		mem: memory{
			movingGenerations: []*generation.Generation{
				generation.NewGeneration(true), // young
				generation.NewGeneration(true), // middle
				generation.NewGeneration(true), // old
			},
			largeObjectGeneration: generation.NewGeneration(false),
		},
	}
}
