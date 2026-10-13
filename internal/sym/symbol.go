package sym

import (
	"fmt"
	"mk/internal/types"
	"mk/pkg/util"
)

type SymbolKind int

const SymbolTableKey = "SymbolTable"

const (
	KindVariable SymbolKind = iota // 变量
	KindConstant                   // 常量
	KindFunction                   // 函数
	KindParam                      // 函数参数
	KindType                       // 类型
	KindField                      // 字段
	KindMethod                     // 实例方法
	KindBuiltin                    // 内置函数/内置类型
	KindModule                     // 模块
)

func (k SymbolKind) String() string {
	switch k {
	case KindVariable:
		return "variable"
	case KindConstant:
		return "constant"
	case KindFunction:
		return "function"
	case KindParam:
		return "parameter"
	case KindType:
		return "type"
	case KindField:
		return "field"
	case KindMethod:
		return "method"
	case KindBuiltin:
		return "builtin"
	case KindModule:
		return "module"
	default:
		return "unknown"
	}
}

type Symbol struct {
	Name   string
	Kind   SymbolKind
	Type   *types.TypeDef
	Index  int
	Depth  int
	Offset int // 栈帧偏移

	// 源码位置

	File   string
	Line   int
	Column int

	Constant bool
	Extern   bool
	Comments string

	Extra map[string]any
}

func (s *Symbol) String() string {
	if s == nil {
		return "<nil>"
	}
	kind := s.Kind.String()
	typeName := "<none>"
	if s.Type != nil {
		typeName = s.Type.Name
	}
	return fmt.Sprintf("[%s] %s : %s (depth=%d, index=%d)", kind, s.Name, typeName, s.Depth, s.Index)
}

func (s *Symbol) Location() string {
	if s.File == "" {
		return fmt.Sprintf("%d:%d", s.Line, s.Column)
	}
	return fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.Column)
}

func NewRootTable(ctx *util.Context) *SymbolTable {
	if st, ok := ctx.Get(SymbolTableKey).(*SymbolTable); ok {
		return st
	}
	idx := 0
	tab := &SymbolTable{
		name:  "<global>",
		ctx:   ctx,
		store: map[string]*Symbol{},
		depth: 0,
		index: &idx,
		outer: nil,

		owner:  nil,
		locals: []*Symbol{},
	}
	ctx.Set(SymbolTableKey, tab)
	tab.root = tab
	tab.current = tab
	return tab
}

type SymbolTable struct {
	root    *SymbolTable
	current *SymbolTable

	name string
	ctx  *util.Context

	store map[string]*Symbol
	depth int
	index *int
	outer *SymbolTable

	owner  *Symbol
	locals []*Symbol
}

func (tab *SymbolTable) Enter(name string) *SymbolTable {
	child := &SymbolTable{
		root:    tab.root,
		current: nil,

		name: name,
		ctx:  tab.current.ctx,

		store: map[string]*Symbol{},
		depth: tab.current.depth + 1,
		index: tab.current.index,
		outer: tab.current,

		owner:  tab.current.owner,
		locals: []*Symbol{},
	}
	tab.current = child
	return tab
}

func (tab *SymbolTable) Leave() *SymbolTable {
	if tab.current.outer != nil {
		tab.current = tab.current.outer
	}
	return tab
}

func (tab *SymbolTable) Index() int {
	return *tab.current.index
}

func (tab *SymbolTable) Depth() int {
	return tab.current.depth
}

func (tab *SymbolTable) Define(name string, kind SymbolKind, typ *types.TypeDef) (*Symbol, error) {
	cur := tab.current
	if _, exists := cur.store[name]; exists {
		return nil, fmt.Errorf("variable %s already defined in this scope.", name)
	}
	sym := &Symbol{
		Name:  name,
		Kind:  kind,
		Type:  typ,
		Index: *cur.index,
		Depth: cur.depth,
	}
	*cur.index = *cur.index + 1
	cur.store[name] = sym
	cur.locals = append(cur.locals, sym)
	return sym, nil
}

func (tab *SymbolTable) SetOwner(sym *Symbol) {
	tab.current.owner = sym
}

func (tab *SymbolTable) Owner() *Symbol {
	return tab.current.owner
}

func (tab *SymbolTable) Name() string {
	return tab.current.name
}

func (tab *SymbolTable) Parent() *SymbolTable {
	return tab.current.outer
}

func (tab *SymbolTable) Resolve(name string) (*Symbol, bool) {
	for t := tab.current; t != nil; t = t.outer {
		if symbol, ok := t.store[name]; ok {
			return symbol, ok
		}
	}
	return nil, false
}
