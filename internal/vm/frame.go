package vm

import (
	"mk/internal/oop"
	"mk/internal/opcode"
	"mk/internal/sym"
)

// Frame 一次函数调用的栈帧（保留以兼容可能的直接调用场景）。
type Frame struct {
	cl          *oop.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *oop.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: 0, basePointer: basePointer}
}

// Instructions 返回该帧对应函数的指令流。cl.Fn 可能是 *sym.CompiledFunction
// 或 *oop.Builtin，后者没有指令流，返回 nil。
func (f *Frame) Instructions() opcode.Instructions {
	if f.cl == nil {
		return nil
	}
	if cf, ok := f.cl.Fn.(*sym.CompiledFunction); ok {
		return cf.Instructions
	}
	return nil
}

// StackFrame 是解释线程实际使用的执行帧，持有当前指令流与栈基址。
type StackFrame struct {
	cl        *oop.Closure
	bytecodes opcode.Instructions
	pc        int
	base      int
}
