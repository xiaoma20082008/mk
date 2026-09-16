package interp

import (
	"mk/internal/ast"
	"mk/internal/compiler"
	"mk/internal/diagnostics"
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

// Exec 编译并执行一份 AST。
// 编译错误与运行错误都直接以 error 返回；诊断信息由词法/语法阶段写入 reporter。
func (e *VMEngine) Exec(p *ast.Program, r diagnostics.DiagnosticReporter) (oop.Obj, error) {
	if p == nil {
		return nil, nil
	}

	// TODO:
	c := compiler.NewCompilerWithState(e.symbols, nil)
	// 1. 编译：AST -> 字节码。
	constants, bytecodes, err := c.Compile(p)
	if err != nil {
		return nil, err
	}
	e.constants = constants

	// 2. 执行：把字节码交给栈式虚拟机
	m := vm.NewVMWithGlobals(e.globals, constants, bytecodes)
	if err := m.Run(); err != nil {
		return nil, err
	}

	// 3. 保存全局区，供下一行继续使用
	e.globals = m.Globals()
	return m.Result(), nil
}
