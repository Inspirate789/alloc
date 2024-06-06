package alloc

import (
	"arena"
	"github.com/Inspirate789/alloc/internal/generation"
	"runtime"
)

type Getter[T any] interface {
	Get() *T
}

type SliceGetter[T any] interface {
	Get() []T
}

type getter[T any] struct {
	metadata *generation.ObjectMetadata
	get      func() *T
}

func (g getter[T]) Get() *T {
	return g.get()
}

type sliceGetter[T any] struct {
	metadata *generation.ObjectMetadata
	get      func() []T
}

func (g sliceGetter[T]) Get() []T {
	return g.get()
}

func New[T any]() Getter[T] {
	metadata := allocateObject[T](mainHypervisor.mem)

	g := getter[T]{
		metadata: metadata,
		get: func() *T {
			metadata.RLock()
			res := (*T)(metadata.Address)
			metadata.RUnlock()
			return res
		},
	}

	runtime.SetFinalizer(&g, func(_ *getter[T]) {
		metadata.Finalized.Swap(true)
	})

	return g
}

func MakeSlice[T any](len, cap int) SliceGetter[T] {
	metadata := allocateSlice[T](mainHypervisor.mem, len, cap)

	g := sliceGetter[T]{
		metadata: &metadata.ObjectMetadata,
		get: func() []T {
			//metadata.RLock()
			res := generation.MakeSliceFromPtr[T](metadata.Address, metadata.Len, metadata.Cap)
			//metadata.RUnlock()
			return res
		},
	}

	runtime.SetFinalizer(&g, func(_ *sliceGetter[T]) {
		metadata.Finalized.Swap(true)
	})

	return g
}

func Append[T any](getter SliceGetter[T], elems ...T) {
	appendSlice(mainHypervisor.mem, getter.Get(), elems...)
}

func Clone[T any](getter Getter[T]) T {
	return arena.Clone[T](*getter.Get())
}

func CloneSlice[S ~[]E, E any](getter SliceGetter[E]) S {
	return arena.Clone[S](getter.Get())
}
