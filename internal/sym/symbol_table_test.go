package sym

import "testing"

func TestSymbolTable_DefineAndResolve(t *testing.T) {
	global := NewSymbolTable()

	// 内建函数写入符号表但不占用全局变量槽位
	if sym, ok := global.Resolve("len"); !ok || sym.Scope != BuiltinScope {
		t.Errorf("builtin len should be resolvable, got %+v (ok=%v)", sym, ok)
	}
	if global.NumDefinitions() != 0 {
		t.Errorf("builtins should not consume definitions, got %d", global.NumDefinitions())
	}

	a := global.Define("a")
	if a.Scope != GlobalScope || a.Index != 0 {
		t.Errorf("a = %+v, want {a GLOBAL 0}", a)
	}

	// 重复定义复用同一个槽位
	if again := global.Define("a"); again.Index != a.Index {
		t.Errorf("redefine a = %+v, want index %d", again, a.Index)
	}

	// 内层作用域：局部变量
	fn := NewEnclosedSymbolTable(global)
	b := fn.Define("b")
	if b.Scope != LocalScope || b.Index != 0 {
		t.Errorf("b = %+v, want {b LOCAL 0}", b)
	}
	if fn.Define("c").Index != 1 {
		t.Errorf("c should take slot 1")
	}

	// 内层可以解析到外层（全局）符号
	if sym, ok := fn.Resolve("a"); !ok || sym.Scope != GlobalScope {
		t.Errorf("a from inner scope = %+v (ok=%v), want GLOBAL", sym, ok)
	}
	if _, ok := global.Resolve("b"); ok {
		t.Errorf("outer scope must not see inner locals")
	}
}

func TestSymbolTable_FreeSymbols(t *testing.T) {
	global := NewSymbolTable()
	global.Define("g")

	outer := NewEnclosedSymbolTable(global)
	outer.Define("x")

	inner := NewEnclosedSymbolTable(outer)

	// 解析外层的局部变量 -> 登记为自由变量
	sym, ok := inner.Resolve("x")
	if !ok {
		t.Fatalf("x should be resolvable")
	}
	if sym.Scope != FreeScope || sym.Index != 0 {
		t.Errorf("x = %+v, want {x FREE 0}", sym)
	}
	if len(inner.FreeSymbols) != 1 {
		t.Fatalf("free symbols = %d, want 1", len(inner.FreeSymbols))
	}
	// 自由变量在定义它的作用域里仍是局部变量
	if inner.FreeSymbols[0].Scope != LocalScope {
		t.Errorf("free symbol should reference a LOCAL, got %s", inner.FreeSymbols[0].Scope)
	}

	// 全局变量与内建函数不需要捕获
	if sym, _ := inner.Resolve("g"); sym.Scope != GlobalScope {
		t.Errorf("g = %+v, want GLOBAL", sym)
	}
	if sym, _ := inner.Resolve("len"); sym.Scope != BuiltinScope {
		t.Errorf("len = %+v, want BUILTIN", sym)
	}
}
