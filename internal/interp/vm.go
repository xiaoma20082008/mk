package interp

import (
	"mk/internal/ast"
	"mk/internal/compiler"
	"mk/internal/diagnostics"
	"mk/internal/gc"
	"mk/internal/oop"
	"mk/internal/sym"
	"mk/internal/vm"
)

type VMEngine struct {
	symbols   sym.Scope // 符号表
	constants []oop.Obj // 常量池
	globals   []oop.Obj // 全局区
}

func NewVMEngine() *VMEngine {
	return &VMEngine{
		symbols: sym.NewScope(nil),
	}
}

func (e *VMEngine) Name() string { return "stack-based vm" }

func (e *VMEngine) Kind() Kind { return KindVM }

func (e *VMEngine) Exec(p *ast.Program, r diagnostics.DiagnosticReporter) (oop.Obj, error) {
	if p == nil {
		return nil, nil
	}

	// 1. 编译：AST -> 字节码
	c := compiler.NewCompilerWithState(e.symbols, e.constants)
	constants, bytecodes, err := c.Compile(p)
	if err != nil {
		return nil, err
	}
	e.constants = constants

	mainFn := &sym.CompiledFunction{Instructions: bytecodes, NumLocals: 0, NumParameters: 0}
	mainClosure := oop.NewClosure(mainFn, nil)

	// 2. 使用虚拟机开启线程运行字节码：堆 + 虚拟机 + 收集器 + 线程。
	heap := vm.NewHeap(1 * 1024 * 1024)
	mvm := vm.NewVM()
	mvm.SetCollector(gc.New(heap))

	mvm.Load(mainClosure, constants, e.globals)
	thread := vm.NewThread(mvm)
	mvm.RegisterThread(thread)
	thread.Execute()

	result, runErr := mvm.Result()
	if runErr != nil {
		return nil, runErr
	}

	// 把本轮可能新建的全局变量写回，供下一行 REPL 继续引用。
	e.globals = mvm.Globals()
	return result, nil
}
