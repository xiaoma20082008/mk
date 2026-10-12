package classfile

import (
	"io"
	"mk/internal/oop"
)

type ClassFile struct {
	Magic   uint32
	Version uint32

	ConstantPool []string

	EntryPoint uint32
}

func ReadFrom(r io.Reader) (*ClassFile, error)

func WriteTo(cf *ClassFile, w io.Writer) error

type ClassLoader interface {
	Load(r io.Reader) (*ClassFile, error)

	Link(cf *ClassFile) *oop.ModuleDef
}
