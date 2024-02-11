package alloc

import "sync/atomic"

type Debugger struct {
	bytesAllocated  atomic.Uint64
	freedBytes      atomic.Uint64
	arenasAllocated atomic.Uint64
}

func readAllocStats() {
	// TODO
}
