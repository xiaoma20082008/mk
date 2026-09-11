package mk

type Env struct {
	outer *Env

	store map[string]Obj
}

func (e *Env) PutString(name string, val string) {
}

func (e *Env) PutInt(name string, val int64) {
}

func (e *Env) PutBool(name string, val bool) {
}

func (e *Env) Put(name string, val Obj) {
	e.store[name] = val
}

func (e *Env) Get(name string) Obj {
	obj, ok := e.store[name]
	if ok {
		return obj
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil
}

func NewEnv(p *Env) *Env {
	return &Env{
		outer: p,
		store: map[string]Obj{},
	}
}
