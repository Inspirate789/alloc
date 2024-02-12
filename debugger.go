package alloc

import "sync/atomic"

type debugger struct {
	// lastCC - Unix timestamp
	lastCC           atomic.Int64 // CC - Cycle Collection
	numCC            atomic.Uint64
	bytesAllocated   atomic.Uint64
	bytesFreed       atomic.Uint64
	arenasAllocated  atomic.Uint64
	arenasFreed      atomic.Uint64
	objectsAllocated atomic.Uint64
	objectsFreed     atomic.Uint64
}

type AllocStats struct {
	// lastCC - Unix timestamp
	LastCC           int64 // CC - Cycle Collection
	NumCC            uint64
	BytesAllocated   uint64
	BytesFreed       uint64
	ArenasAllocated  uint64
	ArenasFreed      uint64
	ObjectsAllocated uint64
	ObjectsFreed     uint64
}

var Debugger = debugger{}

func (d *debugger) ReadAllocStats(stats *AllocStats) {
	if stats != nil {
		stats.LastCC = d.lastCC.Load()
		stats.NumCC = d.numCC.Load()
		stats.BytesAllocated = d.bytesAllocated.Load()
		stats.BytesFreed = d.bytesFreed.Load()
		stats.ArenasAllocated = d.arenasAllocated.Load()
		stats.ArenasFreed = d.arenasFreed.Load()
		stats.ObjectsAllocated = d.objectsAllocated.Load()
		stats.ObjectsFreed = d.objectsFreed.Load()
	}
}
