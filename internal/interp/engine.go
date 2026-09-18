package interp

import (
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/oop"
)

// Kind 执行引擎的种类
type Kind string

const (
	// KindTree AST 树遍历解释器：没有中间表示，直接递归求值 AST
	KindTree Kind = "tree"
	// KindVM 栈式字节码虚拟机：先编译成字节码，再由 VM 逐条解释执行
	KindVM Kind = "vm"
)

func (k Kind) String() string { return string(k) }

// Engine 执行引擎的统一抽象：吃一份 AST，吐一个求值结果。
//
// 上层（repl、未来的 lsp / aot）只依赖这个接口，切换后端不必改动调用方。
// 目前有两个实现：
//   - Evaluator（KindTree）：internal/interp/eval.go
//   - VMEngine（KindVM）：internal/interp/vm.go
//
// 后续 IL/IL.md 里那套寄存器式 SSA（MKIL）落地后，可以作为第三个实现直接插进来。
type Engine interface {
	// Name 引擎的可读名字，用于 REPL 提示
	Name() string
	// Kind 引擎种类
	Kind() Kind
	// Exec 执行一份 AST 并返回结果。
	// 运行期错误以 error 返回；词法/语法/语义类诊断写入传入的 DiagnosticReporter。
	// 引擎自身长期持有的状态（全局变量、符号表、环境）由实现各自保存，跨行复用。
	Exec(p *ast.Program, r diagnostics.DiagnosticReporter) (oop.Obj, error)
}
