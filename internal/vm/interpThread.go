package vm

import (
	"fmt"

	"mk/internal/oop"
	"mk/internal/opcode"
	"mk/internal/runtime"
	"mk/internal/sym"
)

// InterpreterThread 虚拟线程：持有自己的调用栈（frames）与操作数栈（stack）。
type InterpreterThread struct {
	id     int32
	status int32
	vm     *VM

	// 1. 调用栈：每一帧对应一次函数调用
	frames []*StackFrame
	index  int

	// 2. 操作数栈
	stack []oop.Obj
	sp    int

	// 执行结果状态
	lastPopped oop.Obj
	returned   bool
}

// 编译期断言：*InterpreterThread 必须满足 runtime.Thread 接口。
var _ runtime.Thread = (*InterpreterThread)(nil)

func (t *InterpreterThread) ID() int32 { return t.id }

func (t *InterpreterThread) Status() int32 { return t.status }

func (t *InterpreterThread) Execute() {
	t.status = runtime.Running
	defer func() { t.status = runtime.Terminated }()

	entry := t.vm.Entry()
	if entry == nil {
		t.vm.SetResult(oop.O_NULL, nil)
		return
	}
	compiledFn, ok := entry.Fn.(*sym.CompiledFunction)
	if !ok {
		t.vm.SetResult(nil, fmt.Errorf("entry is not a compiled function"))
		return
	}

	t.pushFrame(&StackFrame{cl: entry, bytecodes: compiledFn.Instructions, pc: 0, base: 0})

	for t.index >= 0 {
		t.CheckSafepoint()

		frame := t.frames[t.index]
		if frame.pc >= len(frame.bytecodes) {
			// 帧结束，直接退出该帧。
			t.popFrame()
			continue
		}

		op := opcode.Opcode(frame.bytecodes[frame.pc])
		frame.pc++

		def, err := opcode.Lookup(byte(op))
		if err != nil {
			t.fail(fmt.Errorf("execute: %s", err))
			return
		}
		operands, read := opcode.ReadOperands(def, frame.bytecodes[frame.pc:])
		frame.pc += read

		switch op {
		case opcode.OpConstant:
			t.push(t.vm.Constants()[operands[0]])

		case opcode.OpPop:
			t.pop()
		case opcode.OpTrue:
			t.push(oop.O_TRUE)
		case opcode.OpFalse:
			t.push(oop.O_FALSE)
		case opcode.OpNull:
			t.push(oop.O_NULL)
		case opcode.OpDup:
			t.push(t.top())
		case opcode.OpGetGlobal:
			v := t.vm.Globals()[operands[0]]
			if v == nil {
				v = oop.O_NULL
			}
			t.push(v)
		case opcode.OpSetGlobal:
			t.vm.Globals()[operands[0]] = t.pop()
		case opcode.OpGetLocal:
			t.push(t.stack[frame.base+operands[0]])
		case opcode.OpSetLocal:
			t.stack[frame.base+operands[0]] = t.pop()
		case opcode.OpGetFree:
			t.push(frame.cl.Free[operands[0]])
		case opcode.OpSetFree:
			frame.cl.Free[operands[0]] = t.pop()
		case opcode.OpGetBuiltin:
			t.push(oop.Builtins[operands[0]])

		case opcode.OpAdd:
			t.binary(op)
		case opcode.OpSub, opcode.OpMul, opcode.OpDiv, opcode.OpMod,
			opcode.OpShiftLeft, opcode.OpShiftRight:
			t.arith(op)
		case opcode.OpEqual, opcode.OpNotEqual,
			opcode.OpLess, opcode.OpLessEqual, opcode.OpGreater, opcode.OpGreaterEqual:
			t.compare(op)

		case opcode.OpBang:
			t.push(oop.NewBool(!oop.IsTruthy(t.pop())))
		case opcode.OpMinus:
			v, ok := t.pop().(*oop.IntObj)
			if !ok {
				t.fail(fmt.Errorf("type mismatch: operand of '-' is not INT"))
				return
			}
			t.push(oop.NewInt(-v.Value))
		case opcode.OpTilde:
			v, ok := t.pop().(*oop.IntObj)
			if !ok {
				t.fail(fmt.Errorf("type mismatch: operand of '~' is not INT"))
				return
			}
			t.push(oop.NewInt(^v.Value))

		case opcode.OpJump:
			frame.pc = operands[0]
		case opcode.OpJumpNotTruthy:
			if !oop.IsTruthy(t.pop()) {
				frame.pc = operands[0]
			}

		case opcode.OpList:
			count := operands[0]
			elems := make([]oop.Obj, count)
			for i := count - 1; i >= 0; i-- {
				elems[i] = t.pop()
			}
			t.push(t.alloc(16, func() oop.Obj {
				list := oop.NewList()
				for _, e := range elems {
					list.Add(e)
				}
				return list
			}))
		case opcode.OpTuple:
			count := operands[0]
			elems := make([]oop.Obj, count)
			for i := count - 1; i >= 0; i-- {
				elems[i] = t.pop()
			}
			t.push(t.alloc(16, func() oop.Obj { return oop.NewTuple(elems...) }))
		case opcode.OpMap:
			count := operands[0] // 键值对的数量
			m := t.alloc(16, func() oop.Obj { return oop.NewMap() }).(*oop.MapObj)
			for i := 0; i < count; i++ {
				v := t.pop()
				k := t.pop()
				m.Put(k, v)
			}
			t.push(m)

		case opcode.OpIndex:
			right := t.pop()
			left := t.pop()
			t.doIndex(left, right)
			if t.returned {
				return
			}
		case opcode.OpSetIndex:
			value := t.pop()
			index := t.pop()
			collection := t.pop()
			t.doSetIndex(collection, index, value)
			if t.returned {
				return
			}

		case opcode.OpClosure:
			constIdx := operands[0]
			freeCount := operands[1]
			free := make([]oop.Obj, freeCount)
			for i := freeCount - 1; i >= 0; i-- {
				free[i] = t.pop()
			}
			cf, ok := t.vm.Constants()[constIdx].(*sym.CompiledFunction)
			if !ok {
				t.fail(fmt.Errorf("constant %d is not a compiled function", constIdx))
				return
			}
			t.push(oop.NewClosure(cf, free))

		case opcode.OpCall:
			t.call(operands[0])
			if t.returned {
				return
			}

		case opcode.OpReturnValue:
			t.doReturn(t.pop())
		case opcode.OpReturn:
			t.doReturn(oop.O_NULL)

		default:
			t.fail(fmt.Errorf("unsupported opcode: %s", op))
			return
		}
	}

	// 自然结束：返回最后一个被弹出的值（顶层脚本末尾表达式的值）。
	if !t.returned {
		t.vm.SetResult(t.lastPopped, nil)
	}
}

// ------------------------------------------------------------------------------------------
// 调用约定
// ------------------------------------------------------------------------------------------

// call 处理 OpCall：调用栈布局为 [.., callee, arg0, arg1, ..., argN-1]，sp 指向 argN-1 之后。
func (t *InterpreterThread) call(numArgs int) {
	callee := t.stack[t.sp-1-numArgs]
	switch fn := callee.(type) {
	case *oop.Builtin:
		args := make([]oop.Obj, numArgs)
		copy(args, t.stack[t.sp-numArgs:t.sp])
		t.sp = t.sp - numArgs - 1 // 退掉 callee + 所有参数
		result, err := fn.Fn(args...)
		if err != nil {
			t.fail(err)
			return
		}
		t.push(result)

	case *oop.Closure:
		compiledFn, ok := fn.Fn.(*sym.CompiledFunction)
		if !ok {
			t.fail(fmt.Errorf("cannot call %s", fn.Type()))
			return
		}
		basePointer := t.sp - numArgs
		t.pushFrame(&StackFrame{cl: fn, bytecodes: compiledFn.Instructions, pc: 0, base: basePointer})
		t.sp = basePointer + compiledFn.NumLocals

	default:
		t.fail(fmt.Errorf("not a function: %s", callee.Type()))
		return
	}
}

// doReturn 把返回值回填到调用方的栈上，并弹出当前帧。
// 若当前是顶层帧（index 变为 -1），则把返回值记录为最终结果。
func (t *InterpreterThread) doReturn(retval oop.Obj) {
	frame := t.frames[t.index]
	base := frame.base
	t.popFrame()
	if t.index < 0 {
		t.vm.SetResult(retval, nil)
		t.returned = true
		return
	}
	t.stack[base-1] = retval // base-1 即调用前 callee 所在槽位
	t.sp = base
}

// ------------------------------------------------------------------------------------------
// 栈操作
// ------------------------------------------------------------------------------------------

func (t *InterpreterThread) push(obj oop.Obj) {
	if t.sp >= len(t.stack) {
		t.fail(fmt.Errorf("stack overflow"))
		return
	}
	t.stack[t.sp] = obj
	t.sp++
}

func (t *InterpreterThread) pop() oop.Obj {
	if t.sp <= 0 {
		return oop.O_NULL
	}
	t.sp--
	obj := t.stack[t.sp]
	t.stack[t.sp] = nil
	t.lastPopped = obj
	return obj
}

func (t *InterpreterThread) top() oop.Obj {
	if t.sp <= 0 {
		return oop.O_NULL
	}
	return t.stack[t.sp-1]
}

func (t *InterpreterThread) pushFrame(f *StackFrame) {
	if t.index+1 >= len(t.frames) {
		t.frames = append(t.frames, f)
	} else {
		t.frames[t.index+1] = f
	}
	t.index++
}

func (t *InterpreterThread) popFrame() {
	t.frames[t.index] = nil
	t.index--
}

// alloc 通过注册到 VM 的 GC 收集器创建运行时对象；未配置收集器时直接构造。
func (t *InterpreterThread) alloc(size uint32, create func() oop.Obj) oop.Obj {
	if t.vm != nil && t.vm.gc != nil {
		return t.vm.gc.Alloc(size, create)
	}
	return create()
}

func (t *InterpreterThread) fail(err error) {
	t.vm.SetResult(nil, err)
	t.returned = true
	t.index = -1
}

// ------------------------------------------------------------------------------------------
// 二元 / 一元运算
// ------------------------------------------------------------------------------------------

func (t *InterpreterThread) binary(op opcode.Opcode) {
	right := t.pop()
	left := t.pop()
	result, err := t.doAdd(left, right)
	if err != nil {
		t.fail(err)
		return
	}
	t.push(result)
}

func (t *InterpreterThread) doAdd(left, right oop.Obj) (oop.Obj, error) {
	switch l := left.(type) {
	case *oop.IntObj:
		if r, ok := right.(*oop.IntObj); ok {
			return oop.NewInt(l.Value + r.Value), nil
		}
	case *oop.StringObj:
		if r, ok := right.(*oop.StringObj); ok {
			return oop.NewString(l.Value + r.Value), nil
		}
	case *oop.ListObj:
		if r, ok := right.(*oop.ListObj); ok {
			out := oop.NewList()
			for i := 0; i < l.Len(); i++ {
				out.Add(l.Get(i))
			}
			for i := 0; i < r.Len(); i++ {
				out.Add(r.Get(i))
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("type mismatch: %s + %s", left.Type(), right.Type())
}

func (t *InterpreterThread) arith(op opcode.Opcode) {
	right := t.pop()
	left := t.pop()
	l, ok1 := left.(*oop.IntObj)
	r, ok2 := right.(*oop.IntObj)
	if !ok1 || !ok2 {
		t.fail(fmt.Errorf("type mismatch: operands of arithmetic are not both INT"))
		return
	}
	switch op {
	case opcode.OpSub:
		t.push(oop.NewInt(l.Value - r.Value))
	case opcode.OpMul:
		t.push(oop.NewInt(l.Value * r.Value))
	case opcode.OpDiv:
		if r.Value == 0 {
			t.fail(fmt.Errorf("division by zero"))
			return
		}
		t.push(oop.NewInt(l.Value / r.Value))
	case opcode.OpMod:
		if r.Value == 0 {
			t.fail(fmt.Errorf("modulo by zero"))
			return
		}
		t.push(oop.NewInt(l.Value % r.Value))
	case opcode.OpShiftLeft:
		t.push(oop.NewInt(l.Value << uint(r.Value)))
	case opcode.OpShiftRight:
		t.push(oop.NewInt(l.Value >> uint(r.Value)))
	}
}

func (t *InterpreterThread) compare(op opcode.Opcode) {
	right := t.pop()
	left := t.pop()
	switch op {
	case opcode.OpEqual:
		t.push(oop.NewBool(isEqual(left, right)))
		return
	case opcode.OpNotEqual:
		t.push(oop.NewBool(!isEqual(left, right)))
		return
	}
	// 大小比较：支持 int 与 string
	if li, ok := left.(*oop.IntObj); ok {
		if ri, ok := right.(*oop.IntObj); ok {
			t.push(oop.NewBool(cmpInt(op, li.Value, ri.Value)))
			return
		}
	}
	if ls, ok := left.(*oop.StringObj); ok {
		if rs, ok := right.(*oop.StringObj); ok {
			t.push(oop.NewBool(cmpString(op, ls.Value, rs.Value)))
			return
		}
	}
	t.fail(fmt.Errorf("type mismatch: cannot compare %s and %s", left.Type(), right.Type()))
}

func isEqual(a, b oop.Obj) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	switch av := a.(type) {
	case *oop.IntObj:
		if bv, ok := b.(*oop.IntObj); ok {
			return av.Value == bv.Value
		}
	case *oop.BoolObj:
		if bv, ok := b.(*oop.BoolObj); ok {
			return av.Value == bv.Value
		}
	case *oop.NullObj:
		if _, ok := b.(*oop.NullObj); ok {
			return true
		}
	case *oop.StringObj:
		if bv, ok := b.(*oop.StringObj); ok {
			return av.Value == bv.Value
		}
	default:
		return a.Type() == b.Type() && a.Inspect() == b.Inspect()
	}
	return false
}

func cmpInt(op opcode.Opcode, l, r int64) bool {
	switch op {
	case opcode.OpLess:
		return l < r
	case opcode.OpLessEqual:
		return l <= r
	case opcode.OpGreater:
		return l > r
	case opcode.OpGreaterEqual:
		return l >= r
	}
	return false
}

func cmpString(op opcode.Opcode, l, r string) bool {
	switch op {
	case opcode.OpLess:
		return l < r
	case opcode.OpLessEqual:
		return l <= r
	case opcode.OpGreater:
		return l > r
	case opcode.OpGreaterEqual:
		return l >= r
	}
	return false
}

// ------------------------------------------------------------------------------------------
// 索引
// ------------------------------------------------------------------------------------------

func (t *InterpreterThread) doIndex(left, index oop.Obj) {
	switch c := left.(type) {
	case *oop.ListObj:
		if idx, ok := index.(*oop.IntObj); ok {
			t.push(c.Get(int(idx.Value)))
			return
		}
		t.fail(fmt.Errorf("index of LIST must be INT, got %s", index.Type()))
	case *oop.MapObj:
		t.push(c.Get(index))
	case *oop.StringObj:
		if idx, ok := index.(*oop.IntObj); ok {
			if idx.Value >= 0 && idx.Value < int64(len(c.Value)) {
				t.push(oop.NewString(string(c.Value[idx.Value])))
			} else {
				t.push(oop.O_NULL)
			}
			return
		}
		t.fail(fmt.Errorf("index of STRING must be INT, got %s", index.Type()))
	default:
		t.fail(fmt.Errorf("index operator not supported: %s", left.Type()))
	}
}

func (t *InterpreterThread) doSetIndex(collection, index, value oop.Obj) {
	switch c := collection.(type) {
	case *oop.ListObj:
		if idx, ok := index.(*oop.IntObj); ok {
			c.Set(int(idx.Value), value)
			t.push(value)
			return
		}
		t.fail(fmt.Errorf("index of LIST must be INT, got %s", index.Type()))
	case *oop.MapObj:
		c.Put(index, value)
		t.push(value)
	default:
		t.fail(fmt.Errorf("index assignment not supported: %s", collection.Type()))
	}
}

func (t *InterpreterThread) Yield() {}

func (t *InterpreterThread) CheckSafepoint() {}

func (t *InterpreterThread) Suspend() { t.status = runtime.Suspended }
func (t *InterpreterThread) Resume()  { t.status = runtime.Running }

func (t *InterpreterThread) ScanRoots(gc runtime.GarbageCollector) {
	for i := 0; i < t.sp; i++ {
		if t.stack[i] != nil {
			gc.Update(&t.stack[i], t.stack[i])
		}
	}
}
