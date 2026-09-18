package vm

import (
	"mk/internal/oop"
	"mk/internal/opcode"
)

// StackFrame 是解释线程实际使用的执行帧，持有当前指令流与栈基址。
type StackFrame struct {
	cl        *oop.Closure
	bytecodes opcode.Instructions
	pc        int
	base      int
}
