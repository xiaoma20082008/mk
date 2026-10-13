package vm

import (
	"mk/internal/opcode"
	"mk/internal/types"
	"sync/atomic"
	"uuid"
)

var nextThreadID atomic.Uint32

type ThreadState int

const (
	NEW ThreadState = iota
	RUNNABLE
	BLOCKED
	WAITING
	TERMINATED
)

func (s ThreadState) String() string {
	switch s {
	case NEW:
		return "NEW"
	case RUNNABLE:
		return "RUNNABLE"
	case BLOCKED:
		return "BLOCKED"
	case WAITING:
		return "WAITING"
	default:
		return "TERMINATED"
	}
}

type Thread struct {
	id    uint32
	name  string
	state ThreadState

	// 调用栈
	current *stackFrame

	vm *VM
}

func (t *Thread) GetID() uint32 { return t.id }

func (t *Thread) GetName() string { return t.name }

func (t *Thread) SetName(name string) { t.name = name }

func (t *Thread) GetState() ThreadState { return t.state }

func (t *Thread) Execute() error {
	if t.current == nil {
		return nil
	}
	for frame := t.current; frame != nil; frame = frame.caller {
		method := frame.method
		op := method.Code[frame.pc]
		frame.pc++
		switch op {
		case byte(opcode.OpAdd):
			// 1.
			frame.sp--
			v2 := frame.stack[frame.sp]
			frame.sp--
			v1 := frame.stack[frame.sp]
			// 2.
			res := v1 + v2
			// 3.
			frame.stack[frame.sp] = res
			frame.sp++
		case byte(opcode.OpReturn):
			t.popFrame()
		case byte(opcode.OpCall):
		}
	}
	return nil
}

func (t *Thread) pushFrame(method *types.MethodDef) {
}

func (t *Thread) popFrame() *stackFrame {
	if t.current == nil {
		return nil
	}
	f := t.current.caller
	t.current = f
	return f
}

func NewThread(vm *VM) *Thread {
	return NewThreadName(vm, uuid.New().String())
}

func NewThreadName(vm *VM, name string) *Thread {
	id := nextThreadID.Add(1)
	return &Thread{
		id:    id,
		name:  name,
		state: NEW,

		vm: vm,
	}
}
