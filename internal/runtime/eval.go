package runtime

import (
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/oop"
	"mk/internal/pretty"
	"mk/internal/token"
)

type Evaluator struct {
	env *Env
	r   diagnostics.DiagnosticReporter

	ret oop.Obj
	fin bool // 表示是否结束
}

func (e *Evaluator) Eval(p *ast.Program) oop.Obj {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*runtimeError); ok {
				e.ret = oop.O_NULL
				return
			}
			panic(r)
		}
	}()
	p.Accept(e, nil)
	return e.ret
}

func (f *Evaluator) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		if f.fin {
			break
		}
	}
	return n, nil
}

// Expr
func (f *Evaluator) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	if val := f.env.Get(n.Value); val != nil {
		f.ret = val
		return n, nil
	}
	f.r.Report(diagnostics.ErrUndefinedIdentifier, n.Token.Line, n.Token.Column, n.Token.Lit)
	panic(&runtimeError{"undefined vairable: " + n.Token.Lit})
}

func (f *Evaluator) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	m := oop.NewMap()
	for k, v := range n.Value {
		k.Accept(f, x)
		key := f.ret
		v.Accept(f, x)
		val := f.ret
		m.Put(key, val)
	}
	f.ret = m
	return n, nil
}

func (f *Evaluator) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	f.ret = oop.NewString(n.Value)
	return n, nil
}

func (f *Evaluator) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	f.ret = oop.NewInt(n.Value)
	return n, nil
}

func (f *Evaluator) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	ret := oop.NewList()
	for _, expr := range n.Value {
		expr.Accept(f, x)
		ret.Add(f.ret)
	}
	f.ret = ret
	return n, nil
}

func (f *Evaluator) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	vals := []oop.Obj{}
	for _, expr := range n.Value {
		expr.Accept(f, x)
		vals = append(vals, f.ret)
	}
	f.ret = oop.NewTuple(vals...)
	return n, nil
}

func (f *Evaluator) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	f.ret = oop.NewBool(n.Value)
	return n, nil
}

func (f *Evaluator) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	n.Right.Accept(f, 0)
	rhs := f.ret
	switch n.Op.Type {
	case token.PLUS:
		if rhs.Type() != oop.OBJ_INT {
			f.ret = oop.O_NULL
		}
	case token.MINUS:
		if rhs.Type() == oop.OBJ_INT {
			f.ret = oop.NewInt(-rhs.(*oop.IntObj).Value)
		} else {
			f.ret = oop.O_NULL
		}
	case token.BANG:
		if rhs.Type() == oop.OBJ_BOOL {
			if rhs.(*oop.BoolObj).Value {
				f.ret = oop.O_FALSE
			} else {
				f.ret = oop.O_TRUE
			}
		} else {
			f.ret = oop.O_NULL
		}
	case token.TILDE:
		if rhs.Type() == oop.OBJ_INT {
			v := rhs.(*oop.IntObj).Value
			f.ret = oop.NewInt(^v)
		} else {
			f.ret = oop.O_NULL
		}
	}
	return n, nil
}

func (f *Evaluator) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	lhs := f.ret
	n.Rhs.Accept(f, x)
	rhs := f.ret
	if lhs.Type() == oop.OBJ_INT && rhs.Type() == oop.OBJ_INT {
		if (n.Op.Type == token.SLASH || n.Op.Type == token.PERCENT) && rhs.(*oop.IntObj).Value == 0 {
			f.r.Report(diagnostics.ErrDivByZero, n.Op.Line, n.Op.Column, n.Op.Lit)
			panic(&runtimeError{"divided by zero"})
		}
		f.ret = evalIntOp(n.Op.Lit, lhs, rhs)
	} else if n.Op.Type == token.PLUS {
		f.ret = oop.NewString(lhs.Inspect() + rhs.Inspect())
	} else {
		f.ret = oop.O_NULL
	}
	return n, nil
}

func (f *Evaluator) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	n.Fn.Accept(f, x)
	fn, ok := f.ret.(*oop.FuncObj)
	if !ok {
		astFmt := pretty.NewFormatter()
		name := astFmt.Format(n.Fn)
		f.r.Report(diagnostics.ErrNotAFunction, n.Token.Line, n.Token.Column, name)
		panic(&runtimeError{"expect a function, but got: " + n.Token.Lit})
	}
	args := []oop.Obj{}
	for _, arg := range n.Args {
		arg.Accept(f, x)
		args = append(args, f.ret)
	}
	res, done := fn.Call(args)
	f.ret = res
	if done {
		// todo 暂时没想到什么情况下会放外面
		// f.fin = false
	}
	f.fin = false
	return n, nil
}

func (f *Evaluator) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	cond := f.ret
	switch cond {
	case oop.O_TRUE:
		n.Then.Accept(f, x)
	case oop.O_FALSE:
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	return n, nil
}

func (f *Evaluator) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	fo := &oop.FuncObj{}
	definitionEnv := f.env
	fo.Call = func(args []oop.Obj) (oop.Obj, bool) {
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

func (f *Evaluator) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	n.Rhs.Accept(f, x)
	val := f.ret
	switch left := n.Lhs.(type) {
	case *ast.IdentExpr:
		f.env.Put(left.Value, val)
	case *ast.IndexExpr:
		// list[index] = 10
		// tuple[index] = 10
		// map[index] = 10
		left.Lhs.Accept(f, x)
		targetObj := f.ret
		left.Index.Accept(f, x)
		indexObj := f.ret
		switch obj := targetObj.(type) {
		case *oop.ListObj:
			if index, ok := indexObj.(*oop.IntObj); ok {
				obj.Set(int(index.Value), val)
			}
		case *oop.MapObj:
			obj.Put(indexObj, val)
		case *oop.TupleObj:
			if index, ok := indexObj.(*oop.IntObj); ok {
				obj.Set(int(index.Value), val)
			}
		}
	case *ast.DotExpr:
		// x.name = 10
	}
	return n, nil
}

func (f *Evaluator) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	lv := f.ret
	n.Index.Accept(f, x)
	iv := f.ret
	switch obj := lv.(type) {
	case *oop.ListObj:
		if index, ok := iv.(*oop.IntObj); ok {
			f.ret = obj.Get(int(index.Value))
		}
	case *oop.MapObj:
		f.ret = obj.Get(iv)
	case *oop.TupleObj:
		if index, ok := iv.(*oop.IntObj); ok {
			f.ret = obj.Get(int(index.Value))
		}
	}
	return n, nil
}

func (f *Evaluator) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	return n, nil
}

// Stmt
func (f *Evaluator) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	cond := f.ret
	if isTruthy(cond) {
		n.Then.Accept(f, x)
	} else if n.Else != nil {
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	n.Value.Accept(f, x)
	val := f.ret
	f.env.Put(n.Name.Value, val)
	return n, nil
}

func (f *Evaluator) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	for {
		n.Cond.Accept(f, x)
		if f.fin || !isTruthy(f.ret) {
			break
		}
		n.Body.Accept(f, x)
		if f.fin {
			break
		}
	}
	return n, nil
}

func (f *Evaluator) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error) {
	n.Expr.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) {
	n.Value.Accept(f, x)
	f.fin = true
	return n, nil
}

func (f *Evaluator) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		if f.fin {
			break
		}
	}
	return n, nil
}

// Class
func (f *Evaluator) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) { return n, nil }

func NewEvaluator(e *Env, r diagnostics.DiagnosticReporter) *Evaluator {
	return &Evaluator{
		env: e,
		r:   r,
	}
}

func evalIntOp(op string, lhs, rhs oop.Obj) oop.Obj {
	lv := lhs.(*oop.IntObj).Value
	rv := rhs.(*oop.IntObj).Value
	switch op {
	case "+":
		return oop.NewInt(lv + rv)
	case "-":
		return oop.NewInt(lv - rv)
	case "*":
		return oop.NewInt(lv * rv)
	case "/":
		return oop.NewInt(lv / rv)
	case "%":
		return oop.NewInt(lv % rv)
	case ">":
		return oop.NewBool(lv > rv)
	case ">=":
		return oop.NewBool(lv >= rv)
	case "<":
		return oop.NewBool(lv < rv)
	case "<=":
		return oop.NewBool(lv <= rv)
	case "!=":
		return oop.NewBool(lv != rv)
	case "==":
		return oop.NewBool(lv == rv)
	case "<<":
		return oop.NewInt(lv << rv)
	case ">>":
		return oop.NewInt(lv >> rv)
	case "&":
		return oop.NewInt(lv & rv)
	case "|":
		return oop.NewInt(lv | rv)
	default:
		return oop.O_NULL
	}
}

func isTruthy(obj oop.Obj) bool {
	if obj == nil {
		return false
	}
	if _, ok := obj.(*oop.NullObj); ok {
		return false
	}
	if v, ok := obj.(*oop.BoolObj); ok {
		return v.Value
	}
	return true
}

type runtimeError struct {
	msg string
}
