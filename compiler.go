package mk

type Compiler interface {
	Compile(node Node) error
	Bytecode() *Bytecode
}

type compiler struct {
	instructions Instructions
	constants    []Obj
}

type Bytecode struct {
	instructions Instructions
	constants    []Obj
}

func NewCompiler() Compiler {
	return &compiler{
		instructions: Instructions{},
		constants:    []Obj{},
	}
}

func (c *compiler) Compile(node Node) error {
	node.Accept(c, nil)
	return nil
}

func (c *compiler) Bytecode() *Bytecode {
	return &Bytecode{
		instructions: c.instructions,
		constants:    c.constants,
	}
}

func (c *compiler) emit(op Opcode, operands ...int) int {
	return 0
}

func (c *compiler) VisitProgram(n *Program, x any) (Node, error) { return n, nil }

// Expr
func (c *compiler) VisitIndex(n *IndexExpr, x any) (Node, error)      { return n, nil }
func (c *compiler) VisitDot(n *DotExpr, x any) (Node, error)          { return n, nil }
func (c *compiler) VisitString(n *StringLitExpr, x any) (Node, error) { return n, nil }
func (c *compiler) VisitInt(n *IntLitExpr, x any) (Node, error)       { return n, nil }
func (c *compiler) VisitList(n *ListLitExpr, x any) (Node, error)     { return n, nil }
func (c *compiler) VisitTuple(n *TupleLitExpr, x any) (Node, error)   { return n, nil }
func (c *compiler) VisitMap(n *MapLitExpr, x any) (Node, error)       { return n, nil }
func (c *compiler) VisitBool(n *BoolLitExpr, x any) (Node, error)     { return n, nil }
func (c *compiler) VisitIdent(n *IdentExpr, x any) (Node, error)      { return n, nil }
func (c *compiler) VisitUnary(n *UnaryExpr, x any) (Node, error) {
	n.Right.Accept(c, x)
	switch n.Op.Type {
	case PLUS:
	case MINUS:
	case BANG:
	case TILDE:
	}
	return n, nil
}
func (c *compiler) VisitBinary(n *BinaryExpr, x any) (Node, error) {
	// 后缀表达式
	n.Lhs.Accept(c, x)
	n.Rhs.Accept(c, x)
	switch n.Op.Type {
	case PLUS:
		c.emit(OpAdd)
	case MINUS:
		c.emit(OpSub)
	case STAR:
		c.emit(OpMul)
	case SLASH:
		c.emit(OpDiv)
	case PERCENT:
		c.emit(OpMod)
	}
	return n, nil
}
func (c *compiler) VisitCall(n *CallExpr, x any) (Node, error) {
	n.Fn.Accept(c, x)
	for _, arg := range n.Args {
		arg.Accept(c, x)
	}
	c.emit(OpCall, len(n.Args))
	return n, nil
}
func (c *compiler) VisitTernary(n *TernaryExpr, x any) (Node, error) { return n, nil }
func (c *compiler) VisitParen(n *ParenExpr, x any) (Node, error)     { return n, nil }
func (c *compiler) VisitFn(n *FnExpr, x any) (Node, error)           { return n, nil }
func (c *compiler) VisitAssign(n *AssignExpr, x any) (Node, error)   { return n, nil }

// Stmt
func (c *compiler) VisitIf(n *IfStmt, x any) (Node, error)         { return n, nil }
func (c *compiler) VisitLet(n *LetStmt, x any) (Node, error)       { return n, nil }
func (c *compiler) VisitWhile(n *WhileStmt, x any) (Node, error)   { return n, nil }
func (c *compiler) VisitExpr(n *ExprStmt, x any) (Node, error)     { return n, nil }
func (c *compiler) VisitReturn(n *ReturnStmt, x any) (Node, error) { return n, nil }
func (c *compiler) VisitBlock(n *BlockStmt, x any) (Node, error)   { return n, nil }

// Class
func (c *compiler) VisitClass(n *ClassStmt, x any) (Node, error) { return n, nil }
