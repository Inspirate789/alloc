package alloc

import "alloc/generation"

type memory struct {
	youngGeneration       *generation.Generation
	oldGeneration         *generation.Generation
	thirdGeneration       *generation.Generation
	largeObjectGeneration *generation.Generation
}

var mem memory

func init() {
	// TODO
}

func allocateObject[T any]() (get func() *T, finalize func()) {
	// TODO: check if object is large
	// TODO: check if gc is needed
	return generation.AllocateObject[T](mem.youngGeneration)
}

func allocateSlice[T any](len, cap int) (get func() []T, finalize func()) {
	// TODO: check if object is large
	// TODO: check if gc is needed
	return generation.AllocateSlice[T](mem.youngGeneration, len, cap)
}
