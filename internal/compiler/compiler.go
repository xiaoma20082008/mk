package compiler

import (
	"mk/internal/oop"
	"mk/internal/sym"
)

type Compiler struct {
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func NewCompilerWithState(parent *sym.SymbolTable, constants []oop.Obj) *Compiler {
	c := &Compiler{}
	return c
}
