package alloc

import "alloc/generation"

type memory struct {
	youngGeneration       *generation.Generation
	oldGeneration         *generation.Generation
	thirdGeneration       *generation.Generation
	largeObjectGeneration *generation.Generation
}

func allocateObject[T any](mem memory) (get func() *T, finalize func()) {
	// TODO: check if object is large
	// TODO: check if gc is needed ?
	return generation.AllocateObject[T](mem.youngGeneration)
}

func allocateSlice[T any](mem memory, len, cap int) (get func() []T, finalize func()) {
	// TODO: check if object is large
	// TODO: check if gc is needed ?
	return generation.AllocateSlice[T](mem.youngGeneration, len, cap)
}
