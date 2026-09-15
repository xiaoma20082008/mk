package runtime

import (
	"fmt"
	"mk/internal/oop"
	"mk/internal/opcode"
)

const StackSize = 2048

type VM struct {
	constants    []oop.Obj
	instructions opcode.Instructions

	stack []oop.Obj
	sp    int
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := opcode.Opcode(vm.instructions[ip])
		switch op {
		case opcode.OpConstant:
			index := opcode.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}
		case opcode.OpAdd:
			rhs := vm.pop()
			lhs := vm.pop()
			rv := rhs.(*oop.IntObj).Value
			lv := lhs.(*oop.IntObj).Value
			vm.push(oop.NewInt(lv + rv))
		case opcode.OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) push(o oop.Obj) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() oop.Obj {
	v := vm.stack[vm.sp-1]
	// 复用这个位置
	// vm.stack = vm.stack[:vm.sp]
	vm.sp--
	return v
}

func (vm *VM) top() oop.Obj {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func NewVM(bytecode *Bytecode) *VM {
	return &VM{
		// instructions: bytecode.instructions,
		// constants:    bytecode.constants,

		stack: make([]oop.Obj, StackSize),
		sp:    0,
	}
}
