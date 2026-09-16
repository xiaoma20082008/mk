package vm

import (
	"fmt"

	"mk/internal/oop"
	"mk/internal/opcode"
	"mk/internal/runtime"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

// VM 基于栈的解释器：一条求值栈 + 若干调用栈帧，逐条解释执行字节码。
type VM struct {
	constants    []oop.Obj           // 常量池
	instructions opcode.Instructions // 字节码

	stack []oop.Obj
	sp    int // 指向下一个空闲的栈槽位

	globals []oop.Obj

	frames      []*Frame
	framesIndex int

	// lastPopped 最近一次出栈的值，即整段程序的「执行结果」
	lastPopped oop.Obj

	heap *runtime.Heap
}

// NewVMWithGlobals 复用已有的全局变量表，用于 REPL 场景跨行保持状态
func NewVMWithGlobals(globals []oop.Obj, constants []oop.Obj, instructions opcode.Instructions) *VM {
	if globals == nil {
		globals = make([]oop.Obj, GlobalsSize)
	}
	frames := make([]*Frame, MaxFrames)
	return &VM{
		constants:    constants,
		instructions: instructions,

		stack:       make([]oop.Obj, StackSize),
		sp:          0,
		globals:     globals,
		frames:      frames,
		framesIndex: 1,

		heap: runtime.NewHeap(),
	}
}

// Result 返回程序执行的结果（最近一次出栈的值）
func (vm *VM) Result() oop.Obj {
	if vm.lastPopped == nil {
		return oop.O_NULL
	}
	return vm.lastPopped
}

func (vm *VM) Globals() []oop.Obj {
	return vm.globals
}

func (vm *VM) StackTop() oop.Obj {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// Run 逐条解释执行指令，直到主栈帧结束（或发生运行时错误）
func (vm *VM) Run() error {
	for vm.framesIndex > 0 {
		frame := vm.frames[vm.framesIndex-1]
		ins := frame.Instructions()

		if frame.ip >= len(ins) {
			// 主程序指令流自然结束；函数体一定以返回指令结尾，正常不会走到这里
			if vm.framesIndex == 1 {
				return nil
			}
			vm.popFrame()
			continue
		}

		ip := frame.ip
		op := opcode.Opcode(ins[ip])
		def, err := opcode.Lookup(byte(op))
		if err != nil {
			return err
		}

		// 默认前进到 next 条指令；跳转类指令会在 switch 中改写 ip
		width := 1
		for _, w := range def.Width {
			width += w
		}
		frame.ip = ip + width

		switch op {
		case opcode.OpConstant:
			idx := int(opcode.ReadUint16(ins[ip+1:]))
			if idx >= len(vm.constants) {
				return fmt.Errorf("invalid constant index: %d", idx)
			}
			if err := vm.push(vm.constants[idx]); err != nil {
				return err
			}

		case opcode.OpNull:
			if err := vm.push(oop.O_NULL); err != nil {
				return err
			}
		case opcode.OpTrue:
			if err := vm.push(oop.O_TRUE); err != nil {
				return err
			}
		case opcode.OpFalse:
			if err := vm.push(oop.O_FALSE); err != nil {
				return err
			}

		case opcode.OpPop:
			vm.pop()

		case opcode.OpDup:
			if vm.sp == 0 {
				return fmt.Errorf("nothing to duplicate on the stack")
			}
			if err := vm.push(vm.stack[vm.sp-1]); err != nil {
				return err
			}

		case opcode.OpAdd, opcode.OpSub, opcode.OpMul, opcode.OpDiv, opcode.OpMod,
			opcode.OpEqual, opcode.OpNotEqual,
			opcode.OpGreater, opcode.OpGreaterEqual, opcode.OpLess, opcode.OpLessEqual,
			opcode.OpShiftLeft, opcode.OpShiftRight:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}

		case opcode.OpBang:
			if err := vm.push(oop.NewBool(!oop.IsTruthy(vm.pop()))); err != nil {
				return err
			}
		case opcode.OpMinus:
			val, ok := vm.pop().(*oop.IntObj)
			if !ok {
				return fmt.Errorf("unsupported type for -: expected INT")
			}
			if err := vm.push(oop.NewInt(-val.Value)); err != nil {
				return err
			}
		case opcode.OpTilde:
			val, ok := vm.pop().(*oop.IntObj)
			if !ok {
				return fmt.Errorf("unsupported type for ~: expected INT")
			}
			if err := vm.push(oop.NewInt(^val.Value)); err != nil {
				return err
			}

		case opcode.OpJump:
			frame.ip = int(opcode.ReadUint16(ins[ip+1:]))

		case opcode.OpJumpNotTruthy:
			if !oop.IsTruthy(vm.pop()) {
				frame.ip = int(opcode.ReadUint16(ins[ip+1:]))
			}

		case opcode.OpSetGlobal:
			idx := int(opcode.ReadUint16(ins[ip+1:]))
			if idx >= len(vm.globals) {
				if err := vm.growGlobals(idx + 1); err != nil {
					return err
				}
			}
			vm.globals[idx] = vm.pop()

		case opcode.OpGetGlobal:
			idx := int(opcode.ReadUint16(ins[ip+1:]))
			if idx >= len(vm.globals) {
				return fmt.Errorf("invalid global index: %d", idx)
			}
			val := vm.globals[idx]
			if val == nil {
				val = oop.O_NULL
			}
			if err := vm.push(val); err != nil {
				return err
			}

		case opcode.OpSetLocal:
			idx := int(opcode.ReadUint8(ins[ip+1:]))
			if frame.basePointer+idx >= StackSize {
				return fmt.Errorf("invalid local index: %d", idx)
			}
			vm.stack[frame.basePointer+idx] = vm.pop()

		case opcode.OpGetLocal:
			idx := int(opcode.ReadUint8(ins[ip+1:]))
			if frame.basePointer+idx >= StackSize {
				return fmt.Errorf("invalid local index: %d", idx)
			}
			val := vm.stack[frame.basePointer+idx]
			if val == nil {
				val = oop.O_NULL
			}
			if err := vm.push(val); err != nil {
				return err
			}

		case opcode.OpSetFree:
			idx := int(opcode.ReadUint8(ins[ip+1:]))
			if idx >= len(frame.cl.Free) {
				return fmt.Errorf("invalid free variable index: %d", idx)
			}
			frame.cl.Free[idx] = vm.pop()

		case opcode.OpGetFree:
			idx := int(opcode.ReadUint8(ins[ip+1:]))
			if idx >= len(frame.cl.Free) {
				return fmt.Errorf("invalid free variable index: %d", idx)
			}
			if err := vm.push(frame.cl.Free[idx]); err != nil {
				return err
			}

		case opcode.OpGetBuiltin:
			idx := int(opcode.ReadUint8(ins[ip+1:]))
			if idx >= len(oop.Builtins) {
				return fmt.Errorf("invalid builtin index: %d", idx)
			}
			if err := vm.push(oop.Builtins[idx]); err != nil {
				return err
			}

		case opcode.OpNew:
			// todo 实现new指令,方便后期实现GC
			// vm.heap.Collector().New()
		case opcode.OpCall:
			numArgs := int(opcode.ReadUint8(ins[ip+1:]))
			if vm.sp-1-numArgs < 0 {
				return fmt.Errorf("not enough arguments on the stack for call")
			}
			switch callee := vm.stack[vm.sp-1-numArgs].(type) {
			case *oop.Closure:
				if err := vm.callClosure(callee, numArgs); err != nil {
					return err
				}
			case *oop.Builtin:
				if err := vm.callBuiltin(callee, numArgs); err != nil {
					return err
				}
			default:
				return fmt.Errorf("calling a non-function value")
			}

		case opcode.OpReturnValue:
			value := vm.pop()
			frame := vm.popFrame()
			if vm.framesIndex == 0 {
				// 顶层的 return 直接结束程序
				vm.lastPopped = value
				return nil
			}
			vm.sp = frame.basePointer - 1
			if err := vm.push(value); err != nil {
				return err
			}

		case opcode.OpReturn:
			frame := vm.popFrame()
			if vm.framesIndex == 0 {
				vm.lastPopped = oop.O_NULL
				return nil
			}
			vm.sp = frame.basePointer - 1
			if err := vm.push(oop.O_NULL); err != nil {
				return err
			}

		case opcode.OpClosure:

		case opcode.OpList:
			n := int(opcode.ReadUint16(ins[ip+1:]))
			list := oop.NewList()
			if err := vm.collect(vm.sp, n, func(o oop.Obj) { list.Add(o) }); err != nil {
				return err
			}
			vm.sp -= n
			if err := vm.push(list); err != nil {
				return err
			}

		case opcode.OpTuple:
			n := int(opcode.ReadUint16(ins[ip+1:]))
			values := make([]oop.Obj, 0, n)
			if err := vm.collect(vm.sp, n, func(o oop.Obj) { values = append(values, o) }); err != nil {
				return err
			}
			vm.sp -= n
			if err := vm.push(oop.NewTuple(values...)); err != nil {
				return err
			}

		case opcode.OpMap:
			n := int(opcode.ReadUint16(ins[ip+1:]))
			m := oop.NewMap()
			keys := make([]oop.Obj, 0, n/2)
			values := make([]oop.Obj, 0, n/2)
			if err := vm.collect(vm.sp, n, func(o oop.Obj) {
				if len(keys) == len(values) {
					keys = append(keys, o)
				} else {
					values = append(values, o)
				}
			}); err != nil {
				return err
			}
			vm.sp -= n
			for i := range keys {
				m.Put(keys[i], values[i])
			}
			if err := vm.push(m); err != nil {
				return err
			}

		case opcode.OpIndex:
			index := vm.pop()
			left := vm.pop()
			val, err := vm.evalIndexExpr(left, index)
			if err != nil {
				return err
			}
			if err := vm.push(val); err != nil {
				return err
			}

		case opcode.OpSetIndex:
			value := vm.pop()
			index := vm.pop()
			target := vm.pop()
			if err := vm.assignIndexExpr(target, index, value); err != nil {
				return err
			}
			// 下标赋值本身也是一个表达式，把写入的值留在栈顶
			if err := vm.push(value); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown opcode: %s", op)
		}
	}
	return nil
}

// ------------------------------------------------------------------------------------------
// 求值辅助
// ------------------------------------------------------------------------------------------

func (vm *VM) executeBinaryOperation(op opcode.Opcode) error {
	rhs := vm.pop()
	lhs := vm.pop()

	switch op {
	case opcode.OpEqual:
		return vm.push(oop.NewBool(equalObjs(lhs, rhs)))
	case opcode.OpNotEqual:
		return vm.push(oop.NewBool(!equalObjs(lhs, rhs)))
	}

	li, lhsIsInt := lhs.(*oop.IntObj)
	ri, rhsIsInt := rhs.(*oop.IntObj)
	if lhsIsInt && rhsIsInt {
		return vm.executeIntegerBinaryOperation(op, li.Value, ri.Value)
	}

	if op == opcode.OpAdd {
		ls, lhsIsStr := lhs.(*oop.StringObj)
		rs, rhsIsStr := rhs.(*oop.StringObj)
		switch {
		case lhsIsStr && rhsIsStr:
			return vm.push(oop.NewString(ls.Value + rs.Value))
		case lhsIsStr:
			return vm.push(oop.NewString(ls.Value + rhs.Inspect()))
		case rhsIsStr:
			return vm.push(oop.NewString(lhs.Inspect() + rs.Value))
		}
	}

	return fmt.Errorf("unsupported types for %s: %s %s", op, lhs.Type(), rhs.Type())
}

func (vm *VM) executeIntegerBinaryOperation(op opcode.Opcode, lv, rv int64) error {
	switch op {
	case opcode.OpAdd:
		return vm.push(oop.NewInt(lv + rv))
	case opcode.OpSub:
		return vm.push(oop.NewInt(lv - rv))
	case opcode.OpMul:
		return vm.push(oop.NewInt(lv * rv))
	case opcode.OpDiv:
		if rv == 0 {
			return fmt.Errorf("divided by zero")
		}
		return vm.push(oop.NewInt(lv / rv))
	case opcode.OpMod:
		if rv == 0 {
			return fmt.Errorf("divided by zero")
		}
		return vm.push(oop.NewInt(lv % rv))
	case opcode.OpEqual:
		return vm.push(oop.NewBool(lv == rv))
	case opcode.OpNotEqual:
		return vm.push(oop.NewBool(lv != rv))
	case opcode.OpGreater:
		return vm.push(oop.NewBool(lv > rv))
	case opcode.OpGreaterEqual:
		return vm.push(oop.NewBool(lv >= rv))
	case opcode.OpLess:
		return vm.push(oop.NewBool(lv < rv))
	case opcode.OpLessEqual:
		return vm.push(oop.NewBool(lv <= rv))
	case opcode.OpShiftLeft:
		return vm.push(oop.NewInt(lv << rv))
	case opcode.OpShiftRight:
		return vm.push(oop.NewInt(lv >> rv))
	}
	return fmt.Errorf("unknown integer operator: %s", op)
}

func (vm *VM) evalIndexExpr(left, index oop.Obj) (oop.Obj, error) {
	switch {
	case left.Type() == oop.OBJ_LIST && index.Type() == oop.OBJ_INT:
		list := left.(*oop.ListObj)
		return getSequence(list, int(index.(*oop.IntObj).Value))
	case left.Type() == oop.OBJ_TUPLE && index.Type() == oop.OBJ_INT:
		tuple := left.(*oop.TupleObj)
		return getSequence(tuple, int(index.(*oop.IntObj).Value))
	case left.Type() == oop.OBJ_MAP:
		return left.(*oop.MapObj).Get(index), nil
	}
	return nil, fmt.Errorf("index operator not supported: %s[%s]", left.Type(), index.Type())
}

func (vm *VM) assignIndexExpr(target, index, value oop.Obj) error {
	switch {
	case target.Type() == oop.OBJ_LIST && index.Type() == oop.OBJ_INT:
		return setSequence(target.(*oop.ListObj), int(index.(*oop.IntObj).Value), value)
	case target.Type() == oop.OBJ_TUPLE && index.Type() == oop.OBJ_INT:
		return setSequence(target.(*oop.TupleObj), int(index.(*oop.IntObj).Value), value)
	case target.Type() == oop.OBJ_MAP:
		target.(*oop.MapObj).Put(index, value)
		return nil
	}
	return fmt.Errorf("index assignment not supported: %s[%s]", target.Type(), index.Type())
}

type sequence interface {
	Len() int
}

func getSequence(s sequence, i int) (oop.Obj, error) {
	if i < 0 || i >= s.Len() {
		return nil, fmt.Errorf("index out of bounds: %d", i)
	}
	switch obj := s.(type) {
	case *oop.ListObj:
		return obj.Get(i), nil
	case *oop.TupleObj:
		return obj.Get(i), nil
	}
	return oop.O_NULL, nil
}

func setSequence(s sequence, i int, value oop.Obj) error {
	if i < 0 || i >= s.Len() {
		return fmt.Errorf("index out of bounds: %d", i)
	}
	switch obj := s.(type) {
	case *oop.ListObj:
		obj.Set(i, value)
	case *oop.TupleObj:
		obj.Set(i, value)
	}
	return nil
}

// ------------------------------------------------------------------------------------------
// 调用约定
// ------------------------------------------------------------------------------------------

func (vm *VM) callClosure(cl *oop.Closure, numArgs int) error {
	return nil
}

func (vm *VM) callBuiltin(b *oop.Builtin, numArgs int) error {
	args := make([]oop.Obj, numArgs)
	copy(args, vm.stack[vm.sp-numArgs:vm.sp])

	result, err := b.Fn(args...)
	vm.sp = vm.sp - numArgs - 1
	if err != nil {
		return err
	}
	if result == nil {
		result = oop.O_NULL
	}
	return vm.push(result)
}

// ------------------------------------------------------------------------------------------
// 栈操作
// ------------------------------------------------------------------------------------------

func (vm *VM) push(o oop.Obj) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() oop.Obj {
	if vm.sp == 0 {
		return nil
	}
	v := vm.stack[vm.sp-1]
	vm.sp--
	vm.lastPopped = v
	return v
}

func (vm *VM) top() oop.Obj {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) collect(sp, n int, fn func(oop.Obj)) error {
	if sp-n < 0 {
		return fmt.Errorf("not enough elements on the stack")
	}
	for i := sp - n; i < sp; i++ {
		fn(vm.stack[i])
	}
	return nil
}

func (vm *VM) growGlobals(size int) error {
	if size > GlobalsSize {
		return fmt.Errorf("too many global variables")
	}
	for len(vm.globals) < size {
		vm.globals = append(vm.globals, nil)
	}
	return nil
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// equalObjs 值相等比较：整数/字符串按值，布尔/null 是单例可直接比指针
func equalObjs(lhs, rhs oop.Obj) bool {
	if lhs == nil || rhs == nil {
		return lhs == rhs
	}
	switch left := lhs.(type) {
	case *oop.IntObj:
		if right, ok := rhs.(*oop.IntObj); ok {
			return left.Value == right.Value
		}
	case *oop.StringObj:
		if right, ok := rhs.(*oop.StringObj); ok {
			return left.Value == right.Value
		}
	}
	return lhs == rhs
}
