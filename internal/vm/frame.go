package vm

import "mk/internal/types"

type Value uint64

type stackFrame struct {
	module *types.ModuleDef
	method *types.MethodDef

	caller *stackFrame

	base int
	pc   int

	locals []Value
	stack  []Value
	sp     int
}

func (f *stackFrame) Push(val Value) {
	f.stack[f.sp] = val
	f.sp++
}

func (f *stackFrame) Pop() Value {
	f.sp--
	return f.stack[f.sp]
}

func newFrame(mod *types.ModuleDef) *stackFrame {
	return &stackFrame{
		module: mod,
		method: mod.EntryPoint,

		caller: nil,

		base: 0,
		pc:   0,

		locals: make([]Value, 0),
		stack:  make([]Value, 0),
		sp:     0,
	}
}
