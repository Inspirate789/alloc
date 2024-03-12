package alloc

type hypervisor struct {
	mem memory
}

var mainHypervisor hypervisor

func init() {
	// TODO
}

// choose when gc is needed
