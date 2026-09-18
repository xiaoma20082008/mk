package analyzer

import (
	"mk/internal/ast"
	"mk/internal/diagnostics"
)

type Analyzer interface {
	Resolve(p *ast.Program)
}

type analyzer struct {
	r diagnostics.DiagnosticReporter
}

func New() Analyzer {
	return &analyzer{}
}

func (a *analyzer) Resolve(p *ast.Program) {
}
