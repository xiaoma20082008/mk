package runtime

import (
	"context"
	"mk/internal/oop"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

type VM struct {
	heap *Heap

	gc  *GarbageCollector
	stw bool

	threads map[int32]*InterpreterThread
	err     error
}

func (vm *VM) RequestSafepoint() {}

func (vm *VM) ReleaseSafepoint() {}

func (vm *VM) Start() {}

func (vm *VM) StopTheWorld() {}

func (vm *VM) ScanThreadRoots(gc GarbageCollector) {
}

func (vm *VM) StartMainThread(ctx context.Context, method *oop.MethodDef) {
	thread := NewInterpreterThread(vm)
	vm.threads[thread.id] = thread
	thread.Execute(ctx, method)
}

var MVM *VM

func New() *VM {
	MVM = &VM{
		heap: NewHeap(0),

		gc: nil,

		threads: map[int32]*InterpreterThread{},
	}
	return MVM
}
