package codegen

import (
	"fmt"
	"mk/internal/ast"
)

type Codegen struct {
}

func (c *Codegen) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	return n, nil
}

// ------------------------------------------------------------------------------------------
// Expr
// ------------------------------------------------------------------------------------------

func (c *Codegen) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitCtx(n *ast.CtxExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	return n, nil
}

func (c *Codegen) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	return nil, fmt.Errorf("fn is not supported yet")
}

func (c *Codegen) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	return nil, fmt.Errorf("member access `.` is not supported yet")
}

// Class
func (c *Codegen) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) {
	return nil, fmt.Errorf("class declaration is not supported yet")
}
