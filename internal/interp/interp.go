package interp

import (
	"mk/internal/runtime"
)

// New 按种类创建执行引擎。
// 1. 树遍历
// 2. VM执行
func New(kind Kind, env *runtime.Env) Engine {
	if kind == KindVM {
		return NewVMEngine()
	}
	return NewEvaluator(env, nil)
}
