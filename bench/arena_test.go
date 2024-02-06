package arena_test

import (
	"arena"
	"testing"
)

type Data struct {
	arr [4096]byte
}

var (
	mem  *arena.Arena
	data *Data
	// gcStats debug.GCStats
)

func init() {
	mem = arena.NewArena()
}

func NewDataNoGC() *Data {
	return arena.New[Data](mem)
}

func NewDataGC() *Data {
	return &Data{}
}

func BenchmarkNoGC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data = NewDataNoGC()
		//b.StopTimer()
		//debug.ReadGCStats(&gcStats)
		//b.Log(gcStats.NumGC, gcStats.PauseTotal)
		//b.StartTimer()
	}
	// mem.Free()
}

func BenchmarkGC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data = NewDataGC()
		//b.StopTimer()
		//debug.ReadGCStats(&gcStats)
		//b.Log(gcStats.NumGC, gcStats.PauseTotal)
		//b.StartTimer()
	}
	// runtime.GC()
}

//func BenchmarkNoGCMultipleArenas(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		ar := arena.NewArena()
//		data = arena.New[Data](ar)
//		// ar.Free()
//	}
//}

//func TestSize(t *testing.T) {
//	println(unsafe.Sizeof(Data{}))
//	println(reflect.TypeOf(Data{}).Size())
//}
