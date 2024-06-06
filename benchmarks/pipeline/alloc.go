package main

import (
	"crypto/rand"
	"github.com/Inspirate789/alloc"
)

var resBufAlloc []byte

func allocatorAlloc(bufSize, count int) <-chan alloc.SliceGetter[byte] {
	output := make(chan alloc.SliceGetter[byte])
	go func() {
		for range count {
			output <- alloc.MakeSlice[byte](bufSize/count, bufSize/count)
		}

		close(output)
	}()

	return output
}

func writerAlloc(input <-chan alloc.SliceGetter[byte]) <-chan alloc.SliceGetter[byte] {
	output := make(chan alloc.SliceGetter[byte])
	go func() {
		for buf := range input {
			rand.Read(buf.Get())
			output <- buf
		}

		close(output)
	}()

	return output
}

func readerAlloc(input <-chan alloc.SliceGetter[byte]) <-chan alloc.SliceGetter[byte] {
	output := make(chan alloc.SliceGetter[byte])
	go func() {
		for buf := range input {
			resBufAlloc = buf.Get()
			output <- buf
		}

		close(output)
	}()

	return output
}

func main() {
	for buf := range readerAlloc(writerAlloc(allocatorAlloc(100, 100))) {
		println(buf.Get()[0])
	}
}
