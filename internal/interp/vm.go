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

// VMEngine 栈式字节码虚拟机引擎。
// 链路是 compile（AST -> bytecode）再交给 vm 逐条解释执行。
//
// 有三份状态需要跨行复用：
//   - symbols：符号表，保证上一行定义的名字下一行还能解析到同一个槽位
//   - constants：常量池，保证函数体里的 OpConstant 下标不会跨行错位
//   - globals：全局区，保存顶层变量与函数的运行期值
type VMEngine struct {
	symbols   *sym.SymbolTable
	constants []oop.Obj
	globals   []oop.Obj
}

func NewVMEngine() *VMEngine {
	return &VMEngine{
		symbols: sym.NewSymbolTable(),
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
