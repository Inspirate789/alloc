package arena_test

import (
	"testing"
	"unsafe"
)

var a int = 1

func BenchmarkSizeofNoAlloc(b *testing.B) {
	var size uintptr
	for i := 0; i < b.N; i++ {
		size = unsafe.Sizeof(a)
	}
	_ = size
}

func BenchmarkSizeofAlloc(b *testing.B) {
	var size uintptr
	for i := 0; i < b.N; i++ {
		size = unsafe.Sizeof(*new(int)) // 0 allocs/op !!!!!
	}
	_ = size
}
