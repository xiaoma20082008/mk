package compiler

import (
	"io"
	"os"
)

type FileObject interface {
	Kind() string
	Name() string
	OpenReader() (io.Reader, error)
	OpenWriter() (io.Writer, error)
}

type MemoryFileObject struct{}

func (o *MemoryFileObject) Kind() string
func (o *MemoryFileObject) Name() string
func (o *MemoryFileObject) OpenReader() (io.Reader, error)
func (o *MemoryFileObject) OpenWriter() (io.Writer, error)

type PathFileObject struct {
	path string
}

func (o *PathFileObject) Kind() string {
	return "Source"
}

func (o *PathFileObject) Name() string {
	return o.path
}

func (o *PathFileObject) OpenReader() (io.Reader, error) {
	return os.Open(o.path)
}

func (o *PathFileObject) OpenWriter() (io.Writer, error) {
	return os.Create(o.path)
}

type FileObjectManager struct{}

func NewMemoryFile() FileObject {
	return &MemoryFileObject{}
}

func NewPathFile(path string) FileObject {
	return &PathFileObject{
		path: path,
	}
}
