package sym

import "mk/internal/oop"

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

// Symbol 编译期解析出来的名字绑定
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable 编译期的符号表。函数作用域会形成一条 outer 链：
// 在内层表里解析到外层函数的局部变量时，该变量会被登记为「自由变量」（FreeScope）。
type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int

	// FreeSymbols 本层（函数）捕获到的自由变量，顺序即 OpClosure 之后的压栈顺序
	FreeSymbols []Symbol

	outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	s := &SymbolTable{
		store:       map[string]Symbol{},
		FreeSymbols: []Symbol{},
	}
	// 内建函数是语言级的名字，任何作用域都能直接引用，因此在最外层表里就注册好。
	// 索引即 Builtins 切片下标，运行期 OpGetBuiltin 按下标取用。
	for i, b := range oop.Builtins {
		s.DefineBuiltin(i, b.Name)
	}
	return s
}

// Outer 返回外层符号表。离开作用域（leaveScope）时需要回到外层表继续编译。
func (s *SymbolTable) Outer() *SymbolTable {
	return s.outer
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.outer = outer
	return s
}

// Define 在当前作用域定义名字。若名字已存在则复用原槽位（保证前向引用拿到同一个下标）。
func (s *SymbolTable) Define(name string) Symbol {
	if sym, ok := s.store[name]; ok {
		return sym
	}
	symbol := Symbol{Name: name, Index: s.numDefinitions}
	if s.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

// Resolve 从当前作用域向外查找名字，必要时把外层局部变量登记为自由变量
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	sym, ok := s.store[name]
	if ok || s.outer == nil {
		return sym, ok
	}

	sym, ok = s.outer.Resolve(name)
	if !ok {
		return sym, false
	}

	// 全局变量和内建函数不随栈帧消失，可以直接引用
	if sym.Scope == GlobalScope || sym.Scope == BuiltinScope {
		return sym, true
	}

	// 外层函数的局部变量 / 自由变量：本层需要通过 Free 列表捕获
	return s.defineFree(sym), true
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1, Scope: FreeScope}
	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) NumDefinitions() int {
	return s.numDefinitions
}
