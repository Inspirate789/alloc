//go:build goexperiment.arenas

package alloc

import (
	"arena"
	"runtime"
)

type Getter[T any] interface {
	Get() *T
}

type SliceGetter[T any] interface {
	Get() []T
}

type getter[T any] struct {
	get func() *T
}

func (g getter[T]) Get() *T {
	return g.get()
}

type sliceGetter[T any] struct {
	get func() []T
}

func (g sliceGetter[T]) Get() []T {
	return g.get()
}

func New[T any]() Getter[T] {
	get, finalize := allocateObject[T]()
	g := getter[T]{get: get}
	runtime.SetFinalizer(&g, func(_ *getter[T]) { finalize() })

	return g
}

func MakeSlice[T any](len, cap int) SliceGetter[T] {
	get, finalize := allocateSlice[T](len, cap)
	g := sliceGetter[T]{get: get}
	runtime.SetFinalizer(&g, func(_ *getter[T]) { finalize() })

	return g
}

func Clone[T any](getter Getter[T]) T {
	return arena.Clone[T](*getter.Get())
}

func CloneSlice[S ~[]E, E any](getter SliceGetter[E]) S {
	return arena.Clone[S](getter.Get())
}
