package util

type Context struct {
	values map[string]any
}

func (ctx *Context) Get(key string) any {
	if ctx == nil || ctx.values == nil {
		return nil
	}
	if v, ok := ctx.values[key]; ok {
		return v
	}
	return nil
}

func (ctx *Context) Set(key string, val any) {
	if ctx.values == nil {
		ctx.values = map[string]any{}
	}
	ctx.values[key] = val
}

func (ctx *Context) Delete(key string) {
	if ctx.values != nil {
		delete(ctx.values, key)
	}
}

func (ctx *Context) Keys() []string {
	keys := make([]string, 0, len(ctx.values))
	for k := range ctx.values {
		keys = append(keys, k)
	}
	return keys
}

func NewContext() *Context {
	return &Context{
		values: map[string]any{},
	}
}
