package mk

import "fmt"

const StackSize = 2048

type VM struct {
	constants    []Obj
	instructions Instructions

	stack []Obj
	sp    int
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := Opcode(vm.instructions[ip])
		switch op {
		case OpConstant:
			index := ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}
		case OpAdd:
			rhs := vm.pop()
			lhs := vm.pop()
			rv := rhs.(*IntObj).Value
			lv := lhs.(*IntObj).Value
			vm.push(NewInt(lv + rv))
		case OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) push(o Obj) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() Obj {
	v := vm.stack[vm.sp-1]
	// 复用这个位置
	// vm.stack = vm.stack[:vm.sp]
	vm.sp--
	return v
}

func (vm *VM) top() Obj {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func NewVM(bytecode *Bytecode) *VM {
	return &VM{
		instructions: bytecode.instructions,
		constants:    bytecode.constants,

		stack: make([]Obj, StackSize),
		sp:    0,
	}
}
