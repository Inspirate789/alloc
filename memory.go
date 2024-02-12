package alloc

type memory struct {
	youngGeneration generation
	oldGeneration   generation
	thirdGeneration generation
}

var mem memory

func init() {
	// TODO
}

func allocateObject[T any]() (func() *T, func()) {
	return allocateGenerationObject[T](&mem.youngGeneration)
}

func allocateSlice[T any]() (func() []T, func()) {
	return allocateGenerationSlice[T](&mem.youngGeneration)
}
