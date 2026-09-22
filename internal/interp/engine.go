package interp

import (
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/oop"
)

// Kind 执行引擎的种类
type Kind string

const (
	// KindTree AST 树遍历解释器
	KindTree Kind = "tree"
	// KindVM 栈式字节码虚拟机
	KindVM Kind = "vm"
)

func (k Kind) String() string { return string(k) }

// Engine 执行引擎
type Engine interface {
	// Name 引擎的可读名字，用于 REPL 提示
	Name() string
	// Kind 引擎种类
	Kind() Kind
	// Exec 执行一份 AST 并返回结果。
	Exec(p *ast.Program, r diagnostics.DiagnosticReporter) (oop.Obj, error)
}
