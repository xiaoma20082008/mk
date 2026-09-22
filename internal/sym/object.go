package sym

import (
	"fmt"

	"mk/internal/oop"
	"mk/internal/opcode"
)

// CompiledFunction 是编译期产出的函数模板：一段字节码 + 局部变量/参数个数。
// 运行期不会直接执行它，而是被包装成 Closure 后压入栈中。
type CompiledFunction struct {
	Instructions  opcode.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() oop.ObjType {
	return oop.OBJ_COMPILED_FUNC
}

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("<fn args=%d locals=%d>", cf.NumParameters, cf.NumLocals)
}

// Code 返回函数体的指令流。
// opcode.Bytecode.String() 通过 opcode.codeCarrier 接口调用它，
// 这样 opcode 无需 import code，避免两个包循环依赖。
func (cf *CompiledFunction) Code() opcode.Instructions {
	return cf.Instructions
}
