//go:build goexperiment.arenas

package alloc

type Getter[T any] interface {
	Get() *T
}

type SliceGetter[T any] interface {
	Get() []T
}

func init() {
	// TODO
}

func New[T any]() Getter[T] {
	return nil // TODO
}

func MakeSlice[T any](len, cap int) SliceGetter[T] {
	return nil // TODO
}

func Clone[T any](getter Getter[T]) T {
	return *new(T) // TODO
}

func CloneSlice[S ~[]E, E any](getter SliceGetter[E]) S {
	return nil // TODO
}
