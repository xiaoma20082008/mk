package analyzer

import (
	"fmt"
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/oop"
	"mk/internal/sym"
	"mk/internal/types"
)

type Analyzer interface {
	Resolve(p *ast.Program) *sym.SymbolTable
}

type analyzer struct {
	reporter     diagnostics.DiagnosticReporter
	builtinTypes map[string]*oop.TypeDef
	returnType   *types.TypeDef
	currentScope *sym.SymbolTable
}

func New(r diagnostics.DiagnosticReporter) Analyzer {
	return &analyzer{
		reporter:     r,
		builtinTypes: oop.InternalTypes,
		currentScope: sym.NewRootTable(nil),
	}
}

func (a *analyzer) Resolve(p *ast.Program) *sym.SymbolTable {
	if _, err := p.Accept(a, nil); err != nil {
		return nil
	}
	return a.currentScope
}

func (a *analyzer) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	for i, stmt := range n.Statements {
		ns, err := stmt.Accept(a, nil)
		if err != nil {
			return nil, err
		}
		n.Statements[i] = ns.(ast.Statement)
	}
	return n, nil
}

func (a *analyzer) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	a.currentScope.Enter("Block")
	defer func() { a.currentScope.Leave() }()

	for i, stmt := range n.Statements {
		ns, err := stmt.Accept(a, nil)
		if err != nil {
			return nil, err
		}
		n.Statements[i] = ns.(ast.Statement)
	}
	return n, nil
}

func (a *analyzer) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	n.Value.Accept(a, nil)
	a.currentScope.Define(n.Text(), sym.KindVariable, a.returnType)
	return n, nil
}

func (a *analyzer) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	if _, err := n.Cond.Accept(a, x); err != nil {
		return nil, err
	}
	return n, nil
}

func (a *analyzer) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	return n, nil
}

// ------------------------------------------------------------------------------------------
// Expr
// ------------------------------------------------------------------------------------------

func (a *analyzer) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	symbol, found := a.currentScope.Resolve(n.Value)
	if !found {
		a.reporter.Report(diagnostics.ErrUndefinedIdentifier, "", n.Token.Line, n.Token.Column, n.Value)
		panic(&analyzingError{fmt.Sprintf(diagnostics.ErrUndefinedIdentifier.Template, n.Value)})
	}
	a.returnType = symbol.Type
	return n, nil
}

func (a *analyzer) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitCtx(n *ast.CtxExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (a *analyzer) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {

	return n, nil
}

func (a *analyzer) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	return nil, fmt.Errorf("member access `.` is not supported yet")
}

// Class
func (a *analyzer) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) {
	return nil, fmt.Errorf("class declaration is not supported yet")
}

type analyzingError struct {
	msg string
}

func isTypeEquals(t1, t2 *oop.TypeDef) bool {
	if t1 == t2 {
		return true
	}
	return t1.Name == t2.Name
}
