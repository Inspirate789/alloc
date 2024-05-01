package generation

import (
	"arena"
	"github.com/Inspirate789/alloc/internal/limited_arena"
	"testing"
)

type Data struct {
	arr [4096]byte
}

var (
	mem           = arena.NewArena()
	data          *Data
	object        allocatedObject[Data, *Data]
	gen1          = NewGeneration(make(chan<- struct{}, 10))
	gen2          = NewGeneration(make(chan<- struct{}, 10))
	gen3          = NewGeneration(make(chan<- struct{}, 10))
	get           func() *Data
	finalize      func()
	metadata      *ObjectMetadata
	sliceMetadata *SliceMetadata
	slice         []byte
)

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

func BenchmarkAllocate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		object = allocate[Data](gen1, limited_arena.New[Data])
	}
}

func BenchmarkAllocateObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		metadata = AllocateObject[Data](gen2)
	}
	//b.Log(unsafe.Sizeof(*metadata))
	//b.Log(unsafe.Sizeof(metadata.typeInfo))
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

func BenchmarkAllocateSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sliceMetadata = AllocateSlice[byte](gen3, 4096, 4096)
	}
	//b.Log(unsafe.Sizeof(*metadata))
	//b.Log(unsafe.Sizeof(metadata.typeInfo))
}

func BenchmarkGCSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice = make([]byte, 4096)
		//b.StopTimer()
		//debug.ReadGCStats(&gcStats)
		//b.Log(gcStats.NumGC, gcStats.PauseTotal)
		//b.StartTimer()
	}
	// runtime.GC()
}
