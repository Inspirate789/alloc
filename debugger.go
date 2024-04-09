package alloc

import "sync/atomic"

type debugger struct {
	// lastCC - Unix timestamp
	lastCC          atomic.Int64 // CC - Cycle Collection
	numCC           atomic.Uint64
	arenasAllocated atomic.Uint64
	arenasFreed     atomic.Int64
}

var Debugger = debugger{}

type AllocStats struct {
	// LastCC - Unix timestamp
	LastCC          int64 // CC - Cycle Collection
	NumCC           uint64
	ArenasAllocated uint64
	ArenasFreed     int64
}

func (d *debugger) ReadAllocStats(stats *AllocStats) {
	if stats != nil {
		stats.LastCC = d.lastCC.Load()
		stats.NumCC = d.numCC.Load()
		stats.ArenasAllocated = d.arenasAllocated.Load()
		stats.ArenasFreed = d.arenasFreed.Load()
	}
}
