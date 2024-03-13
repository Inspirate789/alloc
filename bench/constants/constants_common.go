//go:build !windows && !(ios && arm64) && !wasm

package constants

const (
	IsWindows  = 0
	IsIosArm64 = 0
	IsWasm     = 0
)
