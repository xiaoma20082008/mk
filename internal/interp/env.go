package interp

type Env struct {
	vars  map[string]any
	outer *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{
		vars:  map[string]any{},
		outer: outer,
	}
}

// Define 在当前作用域定义变量
func (e *Env) Define(name string, v any) {
	e.vars[name] = v
}

// Get 沿作用域链查找
func (e *Env) Get(name string) (any, bool) {
	for env := e; env != nil; env = env.outer {
		if v, ok := env.vars[name]; ok {
			return v, ok
		}
	}
	return nil, false
}

// Assign 沿作用域链查找并赋值
func (e *Env) Assign(name string, v any) bool {
	for env := e; env != nil; env = env.outer {
		if _, ok := env.vars[name]; ok {
			env.vars[name] = v
			return true
		}
	}
	return false
}
