package runtime

import (
	"context"
	"mk/internal/oop"
	"sync/atomic"
)

type ThreadState int

var nextThreadID int32

const (
	Created ThreadState = iota
	Running
	Suspended
	Waiting
	Terminated
)

type StackFrame struct {
	method *oop.MethodDef
	pc     int
	base   int

	localVars any // 局部变量表
	opStack   any // 操作数栈
	retAddr   any // 方法返回地址
	libs      any // 动态链接
}

type InterpreterThread struct {
	vm    *VM
	id    int32
	state ThreadState

	// 1. 调用栈
	frames []*StackFrame
	index  int

	// 2. 操作数栈
	stack []uintptr
	sp    int

	entry *oop.MethodDef
}

func (t *InterpreterThread) ID() int32 { return t.id }

func (t *InterpreterThread) Status() ThreadState { return t.state }

func (t *InterpreterThread) Execute(ctx context.Context, module *oop.MethodDef) {
	t.state = Running

	t.index = -1
	t.sp = 0

	t.state = Terminated
}

func (t *InterpreterThread) Yield() {}

func (t *InterpreterThread) CheckSafepoint() {
	if t.vm.stw {
		oldState := t.state
		t.state = Suspended

		t.vm.StopTheWorld()

		t.state = oldState
	}
}

func (t *InterpreterThread) Suspend() { t.state = Suspended }

func (t *InterpreterThread) Resume() { t.state = Running }

func (t *InterpreterThread) ScanRoots(gc *GarbageCollector) {
	// 1. 扫描操作数栈
	for i := 0; i < t.sp; i++ {
		if t.stack[i] != 0 {
			gc.MarkAndRelocate(t.stack[i])
		}
	}
	// 2. 扫描调用栈
	for i := 0; i < t.index; i++ {
		frame := t.frames[i]
		if frame != nil && frame.method != nil {

		}
	}
}

func NewInterpreterThread(vm *VM) *InterpreterThread {
	return &InterpreterThread{
		id:     atomic.AddInt32(&nextThreadID, 1),
		state:  Created,
		vm:     vm,
		stack:  make([]uintptr, StackSize),
		frames: make([]*StackFrame, 0, MaxFrames),
		index:  -1,
	}
}
