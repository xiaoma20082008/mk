package mk

import "strconv"

type Evaluator struct {
	env *Env

	ret Obj
}

func (e *Evaluator) Eval(p *Program) Obj {
	p.Accept(e, nil)
	return e.ret
}

func (f *Evaluator) VisitProgram(n *Program, x any) (Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
	}
	return n, nil
}

// Expr
func (f *Evaluator) VisitIdent(n *IdentExpr, x any) (Node, error) {
	if val := f.env.Get(n.Value); val != nil {
		f.ret = val
	} else {
		f.ret = O_NULL
	}
	return n, nil
}

func (f *Evaluator) VisitLiteral(n *LiteralExpr, x any) (Node, error) {
	switch n.Token.Type {
	case INT:
		v, _ := strconv.ParseInt(n.Token.Lit, 10, 0)
		f.ret = NewInt(v)
	case STRING:
		f.ret = NewString(n.Token.Lit)
	case TRUE:
		f.ret = O_TRUE
	case FALSE:
		f.ret = O_FALSE
	case NULL:
		f.ret = O_NULL
	default:
		f.ret = O_NULL
	}
	return n, nil
}

func (f *Evaluator) VisitUnary(n *UnaryExpr, x any) (Node, error) {
	n.Right.Accept(f, 0)
	rhs := f.ret
	switch n.Token.Type {
	case PLUS:
		if rhs.Type() != OBJ_INT {
			f.ret = O_NULL
		}
	case MINUS:
		if rhs.Type() == OBJ_INT {
			f.ret = NewInt(-rhs.(*IntObj).Value)
		} else {
			f.ret = O_NULL
		}
	case BANG:
		if rhs.Type() == OBJ_BOOL {
			if rhs.(*BoolObj).Value {
				f.ret = O_FALSE
			} else {
				f.ret = O_TRUE
			}
		} else {
			f.ret = O_NULL
		}
	case TILDE:
		if rhs.Type() == OBJ_INT {
			v := rhs.(*IntObj).Value
			f.ret = NewInt(^v)
		} else {
			f.ret = O_NULL
		}
	}
	return n, nil
}

func (f *Evaluator) VisitBinary(n *BinaryExpr, x any) (Node, error) {
	n.Lhs.Accept(f, x)
	lhs := f.ret
	n.Rhs.Accept(f, x)
	rhs := f.ret
	if lhs.Type() == OBJ_INT && rhs.Type() == OBJ_INT {
		f.ret = evalIntOp(n.Op.Lit, lhs, rhs)
	}
	return n, nil
}

func (f *Evaluator) VisitCall(n *CallExpr, x any) (Node, error) {
	return n, nil
}

func (f *Evaluator) VisitTernary(n *TernaryExpr, x any) (Node, error) {
	n.Cond.Accept(f, x)
	cond := f.ret
	switch cond {
	case O_TRUE:
		n.Then.Accept(f, x)
	case O_FALSE:
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitParen(n *ParenExpr, x any) (Node, error) {
	return n, nil
}

func (f *Evaluator) VisitFn(n *FnExpr, x any) (Node, error) {
	return n, nil
}

func (f *Evaluator) VisitAssign(n *AssignExpr, x any) (Node, error) {
	return n, nil
}

// Stmt
func (f *Evaluator) VisitIf(n *IfStmt, x any) (Node, error) {
	n.Cond.Accept(f, x)
	cond := f.ret
	if cond == O_TRUE {
		n.Then.Accept(f, x)
	} else {
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitLet(n *LetStmt, x any) (Node, error) {
	n.Value.Accept(f, x)
	val := f.ret
	f.env.Put(n.Name.Value, val)
	return n, nil
}

func (f *Evaluator) VisitFor(n *ForStmt, x any) (Node, error) {
	if n.Init != nil {
		n.Init.Accept(f, x)
	}
	n.Cond.Accept(f, x)
	if n.Post != nil {
		n.Post.Accept(f, x)
	}
	n.Body.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitExpr(n *ExprStmt, x any) (Node, error) {
	n.Expr.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitReturn(n *ReturnStmt, x any) (Node, error) {
	n.Value.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitBlock(n *BlockStmt, x any) (Node, error) {
	return n, nil
}

// Class
func (f *Evaluator) VisitClass(n *ClassStmt, x any) (Node, error) { return n, nil }

func NewEvaluator(e *Env) *Evaluator {
	return &Evaluator{
		env: e,
	}
}

func evalIntOp(op string, lhs, rhs Obj) Obj {
	lv := lhs.(*IntObj).Value
	rv := rhs.(*IntObj).Value
	switch op {
	case "+":
		return NewInt(lv + rv)
	case "-":
		return NewInt(lv / rv)
	case "*":
		return NewInt(lv * rv)
	case "/":
		return NewInt(lv / rv)
	case "%":
		return NewInt(lv % rv)
	case ">":
		return NewBool(lv > rv)
	case ">=":
		return NewBool(lv >= rv)
	case "<":
		return NewBool(lv < rv)
	case "<=":
		return NewBool(lv <= rv)
	case "!=":
		return NewBool(lv != rv)
	case "==":
		return NewBool(lv == rv)
	case "<<":
		return NewInt(lv << rv)
	case ">>":
		return NewInt(lv >> rv)
	case "&":
		return NewInt(lv & rv)
	case "|":
		return NewInt(lv | rv)
	default:
		return O_NULL
	}
}
