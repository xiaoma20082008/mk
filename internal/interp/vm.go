package interp

import (
	"context"
	"mk/internal/ast"
	"mk/internal/classfile"
	"mk/internal/compiler"
	"mk/internal/diagnostics"
	"mk/internal/oop"
	"mk/internal/runtime"
)

type VMEngine struct {
	mvm      *runtime.VM
	compiler compiler.Compiler
	loader   classfile.ClassLoader
}

func NewVMEngine() *VMEngine {
	return &VMEngine{
		mvm:      runtime.New(),
		compiler: compiler.NewCompiler(),
	}
}

func (e *VMEngine) Name() string { return "stack-based vm" }

func (e *VMEngine) Kind() Kind { return KindVM }

func (e *VMEngine) Exec(p *ast.Program, r diagnostics.DiagnosticReporter) (oop.Obj, error) {
	if p == nil {
		return nil, nil
	}

	// 1.
	cf := e.compiler.Compile(p)
	// 2.
	module := e.loader.Link(cf)
	// 3.
	e.mvm.StartMainThread(context.Background(), module.EntryPoint)

	result, runErr := e.mvm.Result()
	if runErr != nil {
		return nil, runErr
	}

	return result, nil
}
