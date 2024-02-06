package arena_test

import (
	"arena"
	"testing"
)

var (
	extMem    *arena.Arena
	sliceNoGC []byte
	sliceGC   []byte
)

func init() {
	extMem = arena.NewArena()
	sliceNoGC = arena.MakeSlice[byte](extMem, 1, 1)
	sliceGC = make([]byte, 1)
}

func BenchmarkNoGCGrowSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sliceNoGC = append(sliceNoGC, 'a')
		// sliceNoGC[len(sliceNoGC)-1] = 'b'
	}
	b.StopTimer()
	b.Log(len(sliceNoGC), cap(sliceNoGC))
	b.StartTimer()
}

func BenchmarkGCGrowSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sliceGC = append(sliceGC, 'a')
		// sliceGC[len(sliceGC)-1] = 'b'
	}
	b.StopTimer()
	b.Log(len(sliceGC), cap(sliceGC))
	b.StartTimer()
}
