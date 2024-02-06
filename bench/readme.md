```
$ export GOEXPERIMENT=arenas
```

```
$ go test -bench=. -benchtime=2047x -benchmem arena_test.go
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-9300H CPU @ 2.40GHz
BenchmarkNoGC-8             2047                19.04 ns/op            0 B/op          0 allocs/op
BenchmarkGC-8               2047              1197 ns/op            4096 B/op          1 allocs/op
PASS
ok      command-line-arguments  0.009s
```

```
$ go test -bench=. -benchtime=2048x -benchmem arena_test.go
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-9300H CPU @ 2.40GHz
BenchmarkNoGC-8             2048               718.1 ns/op          4096 B/op          0 allocs/op
BenchmarkGC-8               2048               491.0 ns/op          4096 B/op          1 allocs/op
PASS
ok      command-line-arguments  0.009s
```

Reason: 

```go
// https://github.com/golang/go/blob/master/src/runtime/arena.go
const (
	// userArenaChunkBytes is the size of a user arena chunk.
	userArenaChunkBytesMax = 8 << 20
	userArenaChunkBytes    = uintptr(int64(userArenaChunkBytesMax-heapArenaBytes)&(int64(userArenaChunkBytesMax-heapArenaBytes)>>63) + heapArenaBytes) // min(userArenaChunkBytesMax, heapArenaBytes)

	// userArenaChunkPages is the number of pages a user arena chunk uses.
	userArenaChunkPages = userArenaChunkBytes / pageSize

	// userArenaChunkMaxAllocBytes is the maximum size of an object that can
	// be allocated from an arena. This number is chosen to cap worst-case
	// fragmentation of user arenas to 25%. Larger allocations are redirected
	// to the heap.
	userArenaChunkMaxAllocBytes = userArenaChunkBytes / 4
)
```

The memory allocation has exceeded the `userArenaChunkBytes`, so memory is not allocated in the arena (?????).

For this reason, you should not store more than `8^20` bytes in the arena.
