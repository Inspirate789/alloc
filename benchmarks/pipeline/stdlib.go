package main

import (
	"crypto/rand"
)

var resBufStdlib []byte

func allocatorStdlib(bufSize, count int) <-chan []byte {
	output := make(chan []byte)
	go func() {
		for range count {
			output <- make([]byte, bufSize/count)
		}

		close(output)
	}()

	return output
}

func writerStdlib(input <-chan []byte) <-chan []byte {
	output := make(chan []byte)
	go func() {
		for buf := range input {
			rand.Read(buf)
			output <- buf
		}

		close(output)
	}()

	return output
}

func readerStdlib(input <-chan []byte) <-chan []byte {
	output := make(chan []byte)
	go func() {
		for buf := range input {
			resBufStdlib = buf
			output <- buf
		}

		close(output)
	}()

	return output
}

//func main() {
//	for buf := range readerStdlib(writerStdlib(allocatorStdlib(100, 100))) {
//		println(buf[0])
//	}
//}
