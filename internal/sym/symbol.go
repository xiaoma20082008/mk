package sym

import (
	"fmt"
	"mk/internal/oop"
)

type SymbolKind int

const (
	KindVariable SymbolKind = iota // 变量
	KindConstant                   // 常量
	KindFunction                   // 函数
	KindParam                      // 函数参数
	KindBuiltin                    // 内置函数/类型
)

type Symbol struct {
	Name  string
	Kind  SymbolKind
	Type  oop.ObjType
	Index int
	Depth int

	// TODO: 将来用于报错提示
	Line   int
	Column int
}

type Scope interface {
	Define(name string, kind SymbolKind, typ oop.ObjType) (*Symbol, error)
	Resolve(name string) (*Symbol, bool)
	Parent() Scope
	Depth() int
	Index() int
}

func NewScope(parent Scope) Scope {
	s := &symbolTable{
		store: map[string]*Symbol{},
		outer: parent,
		depth: 0,
		index: 0,
	}
	if parent != nil {
		s.depth = parent.Depth() + 1
		s.index = parent.Index()
	}
	return s
}

type symbolTable struct {
	store map[string]*Symbol
	depth int
	index int
	outer Scope
}

func (tab *symbolTable) Index() int {
	return tab.index
}

func (tab *symbolTable) Depth() int {
	return tab.depth
}

func (tab *symbolTable) Define(name string, kind SymbolKind, typ oop.ObjType) (*Symbol, error) {
	if _, exists := tab.store[name]; exists {
		return nil, fmt.Errorf("variable %s already defined in this scope.", name)
	}
	sym := &Symbol{
		Name:  name,
		Kind:  kind,
		Type:  typ,
		Index: tab.index,
		Depth: tab.depth,
	}
	tab.index++
	tab.store[name] = sym
	return sym, nil
}

func (tab *symbolTable) Parent() Scope {
	return tab.outer
}

func (tab *symbolTable) Resolve(name string) (*Symbol, bool) {
	sym, ok := tab.store[name]
	if ok || tab.outer == nil {
		return sym, ok
	}
	return tab.outer.Resolve(name)
}
