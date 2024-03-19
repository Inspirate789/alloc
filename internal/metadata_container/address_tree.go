package metadata_container

import "unsafe"

type address interface {
	Address() unsafe.Pointer
}

// TODO: RB-tree
