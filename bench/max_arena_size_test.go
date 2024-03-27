package arena_test

import (
	"alloc/bench/constants"
	"testing"
)

const (
	// _64bit = 1 on 64-bit systems, 0 on 32-bit systems
	_64bit = 1 << (^uintptr(0) >> 63) / 2

	// heapArenaBytes is the size of a heap arena. The heap
	// consists of mappings of size heapArenaBytes, aligned to
	// heapArenaBytes. The initial heap mapping is one arena.
	//
	// This is currently 64MB on 64-bit non-Windows and 4MB on
	// 32-bit and on Windows. We use smaller arenas on Windows
	// because all committed memory is charged to the process,
	// even if it's not touched. Hence, for processes with small
	// heaps, the mapped arena space needs to be commensurate.
	// This is particularly important with the race detector,
	// since it significantly amplifies the cost of committed
	// memory.
	heapArenaBytes = 1 << logHeapArenaBytes

	// logHeapArenaBytes is log_2 of heapArenaBytes. For clarity,
	// prefer using heapArenaBytes where possible (we need the
	// constant to compute some other constants).
	logHeapArenaBytes = (6+20)*(_64bit*(1-constants.IsWindows)*(1-constants.IsWasm)*(1-constants.IsIosArm64)) + (2+20)*(_64bit*constants.IsWindows) + (2+20)*(1-_64bit) + (2+20)*constants.IsWasm + (2+20)*constants.IsIosArm64

	// userArenaChunkBytes is the size of a user arena chunk.
	userArenaChunkBytesMax = 8 << 20
	userArenaChunkBytes    = uintptr(int64(userArenaChunkBytesMax-heapArenaBytes)&(int64(userArenaChunkBytesMax-heapArenaBytes)>>63) + heapArenaBytes) // min(userArenaChunkBytesMax, heapArenaBytes)

	// userArenaChunkMaxAllocBytes is the maximum size of an object that can
	// be allocated from an arena. This number is chosen to cap worst-case
	// fragmentation of user arenas to 25%. Larger allocations are redirected
	// to the heap.
	userArenaChunkMaxAllocBytes = userArenaChunkBytes / 4
)

func TestConstants(t *testing.T) {
	t.Log(_64bit)
	t.Log(heapArenaBytes)
	t.Log(userArenaChunkBytesMax)
	t.Log(userArenaChunkBytes)
	t.Log(float64(userArenaChunkBytes) / userArenaChunkBytesMax)
	t.Log(userArenaChunkMaxAllocBytes)
}
