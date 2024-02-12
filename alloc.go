//go:build goexperiment.arenas

package alloc

import "runtime"

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
	get, linkFinalizer := allocateObject[T]()
	g := getter[T]{
		get: func() *T {
			return get()
		},
	}
	runtime.SetFinalizer(&g, func(_ *getter[T]) { linkFinalizer() })

	return g // &g ?
}

func MakeSlice[T any](len, cap int) SliceGetter[T] {
	get, linkFinalizer := allocateSlice[T]()
	g := sliceGetter[T]{
		get: func() []T {
			return get()
		},
	}
	runtime.SetFinalizer(&g, func(_ *getter[T]) { linkFinalizer() })

	return g // &g ?
}

func Clone[T any](getter Getter[T]) T {
	return *new(T) // TODO
}

func CloneSlice[S ~[]E, E any](getter SliceGetter[E]) S {
	return nil // TODO
}
