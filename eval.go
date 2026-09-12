package mk

import "strconv"

type Evaluator struct {
	env *Env

	ret Obj
	fin bool // 表示是否结束
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
	} else if n.Op.Type == PLUS {
		f.ret = NewString(lhs.Inspect() + rhs.Inspect())
	} else {
		f.ret = O_NULL
	}
	return n, nil
}

func (f *Evaluator) VisitCall(n *CallExpr, x any) (Node, error) {
	n.Fn.Accept(f, x)
	fn, ok := f.ret.(*FuncObj)
	if !ok {
		panic("not a function")
	}
	args := []Obj{}
	for _, arg := range n.Args {
		arg.Accept(f, x)
		args = append(args, f.ret)
	}
	res, done := fn.Call(args)
	f.ret = res
	if done {
		f.fin = false
	}
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
	fo := &FuncObj{}
	definitionEnv := f.env
	fo.Call = func(args []Obj) (Obj, bool) {
		// 1. 创建临时环境
		localEnv := NewEnv(definitionEnv)
		// 2. 绑定形参和实参
		for i, param := range n.Args {
			localEnv.Put(param.Value, args[i])
		}
		// 3. 备份当前环境
		oldEnv := f.env
		f.env = localEnv

		// 4. 执行函数体
		n.Body.Accept(f, x)

		// 5. 记录执行结果和当前的返回信号
		res := f.ret
		fin := f.fin

		// 6. 恢复
		f.env = oldEnv

		// 7. 返回结果给调用者
		return res, fin
	}
	f.ret = fo
	return n, nil
}

func (f *Evaluator) VisitAssign(n *AssignExpr, x any) (Node, error) {
	n.Rhs.Accept(f, x)
	if id, ok := n.Lhs.(*IdentExpr); ok {
		f.env.Put(id.Value, f.ret)
	}
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

func (f *Evaluator) VisitWhile(n *WhileStmt, x any) (Node, error) {
	for {
		n.Cond.Accept(f, x)
		if !isTruthy(f.ret) {
			break
		}
		n.Body.Accept(f, x)
		if f.fin {
			break
		}
	}
	return n, nil
}

func (f *Evaluator) VisitExpr(n *ExprStmt, x any) (Node, error) {
	n.Expr.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitReturn(n *ReturnStmt, x any) (Node, error) {
	n.Value.Accept(f, x)
	f.fin = true
	return n, nil
}

func (f *Evaluator) VisitBlock(n *BlockStmt, x any) (Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		_, done := f.ret, f.fin
		if done {
			break
		}
	}
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

func isTruthy(obj Obj) bool {
	if obj == nil {
		return false
	}
	if _, ok := obj.(*NullObj); ok {
		return false
	}
	if v, ok := obj.(*BoolObj); ok {
		return v.Value
	}
	return true
}
