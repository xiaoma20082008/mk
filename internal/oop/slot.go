package oop

type RawSlot interface {
	Addr() uintptr
	Bytes() []byte
}
