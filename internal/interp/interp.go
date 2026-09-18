package interp

import (
	"mk/internal/runtime"
)

// New 按种类创建执行引擎。
// env 是树遍历引擎的变量环境；VM 引擎自己维护符号表与全局区，忽略该参数。
//
// 两个引擎都是有状态的：树遍历引擎持有 env，VM 引擎持有符号表与全局区，
// 因此 REPL 里应当长期持有实例，跨行复用，而不是每行新建。
func New(kind Kind, env *runtime.Env) Engine {
	if kind == KindVM {
		return NewVMEngine()
	}
	return NewEvaluator(env, nil)
}
