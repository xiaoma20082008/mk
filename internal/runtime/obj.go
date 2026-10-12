package runtime

import "unsafe"

type ObjectHeader struct {
	MarkWord uint64
	TypePtr  uintptr
}

const HEADER_SIZE = unsafe.Sizeof(ObjectHeader{})
