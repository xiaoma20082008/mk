package runtime

import (
	"mk/internal/ast"
	"mk/internal/oop"
	"mk/internal/opcode"
	"mk/internal/token"
)

type Compiler interface {
	Compile(node ast.Node) error
	Bytecode() *Bytecode
}

type compiler struct {
	instructions opcode.Instructions
	constants    []oop.Obj
}

type Bytecode struct {
	instructions opcode.Instructions
	constants    []oop.Obj
}

func NewCompiler() Compiler {
	return &compiler{
		instructions: opcode.Instructions{},
		constants:    []oop.Obj{},
	}
}

func (c *compiler) Compile(node ast.Node) error {
	node.Accept(c, nil)
	return nil
}

func (c *compiler) Bytecode() *Bytecode {
	return &Bytecode{
		instructions: c.instructions,
		constants:    c.constants,
	}
}

func (c *compiler) emit(op opcode.Opcode, operands ...int) int {
	return 0
}

func (c *compiler) VisitProgram(n *ast.Program, x any) (ast.Node, error) { return n, nil }

// Expr
func (c *compiler) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error)      { return n, nil }
func (c *compiler) VisitDot(n *ast.DotExpr, x any) (ast.Node, error)          { return n, nil }
func (c *compiler) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) { return n, nil }
func (c *compiler) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error)       { return n, nil }
func (c *compiler) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error)     { return n, nil }
func (c *compiler) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error)   { return n, nil }
func (c *compiler) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error)       { return n, nil }
func (c *compiler) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error)     { return n, nil }
func (c *compiler) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error)      { return n, nil }
func (c *compiler) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	n.Right.Accept(c, x)
	switch n.Op.Type {
	case token.PLUS:
	case token.MINUS:
	case token.BANG:
	case token.TILDE:
	}
	return n, nil
}
func (c *compiler) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	// 后缀表达式
	n.Lhs.Accept(c, x)
	n.Rhs.Accept(c, x)
	switch n.Op.Type {
	case token.PLUS:
		c.emit(opcode.OpAdd)
	case token.MINUS:
		c.emit(opcode.OpSub)
	case token.STAR:
		c.emit(opcode.OpMul)
	case token.SLASH:
		c.emit(opcode.OpDiv)
	case token.PERCENT:
		c.emit(opcode.OpMod)
	}
	return n, nil
}
func (c *compiler) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	n.Fn.Accept(c, x)
	for _, arg := range n.Args {
		arg.Accept(c, x)
	}
	c.emit(opcode.OpCall, len(n.Args))
	return n, nil
}
func (c *compiler) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) { return n, nil }
func (c *compiler) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error)     { return n, nil }
func (c *compiler) VisitFn(n *ast.FnExpr, x any) (ast.Node, error)           { return n, nil }
func (c *compiler) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error)   { return n, nil }

// Stmt
func (c *compiler) VisitIf(n *ast.IfStmt, x any) (ast.Node, error)         { return n, nil }
func (c *compiler) VisitLet(n *ast.LetStmt, x any) (ast.Node, error)       { return n, nil }
func (c *compiler) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error)   { return n, nil }
func (c *compiler) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error)     { return n, nil }
func (c *compiler) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) { return n, nil }
func (c *compiler) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error)   { return n, nil }

// Class
func (c *compiler) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) { return n, nil }
