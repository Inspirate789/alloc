package alloc

type hypervisor struct {
	mem memory
}

func newHypervisor(mem memory) hypervisor {
	return hypervisor{mem: mem}
}

// choose when gc is needed
