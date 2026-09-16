package vm

import (
	"mk/internal/oop"
	"mk/internal/opcode"
)

// Frame 一次函数调用的栈帧。
// basePointer 指向该帧第 0 号局部变量的栈位置，参数即紧随其后的前 numArgs 个槽位。
type Frame struct {
	cl          *oop.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *oop.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          0,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() opcode.Instructions {
	return nil
}
