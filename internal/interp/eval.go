package interp

import (
	"fmt"
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/pretty"
	"mk/internal/sym"
	"mk/internal/token"
	"mk/internal/types"
	"strconv"
	"strings"
)

type Evaluator struct {
	errors diagnostics.DiagnosticReporter
	symtab *sym.SymbolTable

	returnType *types.TypeDef
	returnVal  any
	returnErr  any
	finished   bool
}

type funtion func(args []any) (*types.TypeDef, any, error)

func (e *Evaluator) Eval(p *ast.Program) any {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*runtimeError); ok {
				e.returnVal = nil
				return
			}
			panic(r)
		}
	}()
	p.Accept(e, nil)
	return e.returnVal
}

func (f *Evaluator) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		if f.finished {
			break
		}
	}
	return n, nil
}

func (f *Evaluator) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	if sym, ok := f.symtab.Resolve(n.Value); ok {
		f.returnVal = sym
		return n, nil
	}
	f.errors.Report(diagnostics.ErrUndefinedIdentifier, "", n.Token.Line, n.Token.Column, n.Token.Lit)
	panic(&runtimeError{fmt.Sprintf(diagnostics.ErrUndefinedIdentifier.Template, n.Token.Lit)})
}

func (f *Evaluator) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	m := map[any]any{}
	for k, v := range n.Value {
		k.Accept(f, x)
		key := f.returnVal
		v.Accept(f, x)
		val := f.returnVal
		m[key] = val
	}
	f.returnVal = m
	return n, nil
}

func (f *Evaluator) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	f.returnVal = n.Value
	f.returnType = types.SysString
	return n, nil
}

func (f *Evaluator) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	f.returnVal = n.Value
	f.returnType = types.SysInt64
	return n, nil
}

func (f *Evaluator) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	ret := []any{}
	for _, expr := range n.Value {
		expr.Accept(f, x)
		ret = append(ret, f.returnVal)
	}
	f.returnVal = ret
	return n, nil
}

func (f *Evaluator) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	vals := []any{}
	for _, expr := range n.Value {
		expr.Accept(f, x)
		vals = append(vals, f.returnVal)
	}
	f.returnVal = vals
	return n, nil
}

func (f *Evaluator) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	f.returnVal = n.Value
	return n, nil
}

func (f *Evaluator) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	n.Right.Accept(f, x)
	rhs := f.returnVal
	switch n.Op.Type {
	case token.PLUS:
		if _, ok := rhs.(int64); !ok {
			f.returnVal = nil
		}
	case token.MINUS:
		if val, ok := rhs.(int64); !ok {
			f.returnVal = -val
		} else {
			f.returnVal = nil
		}
	case token.BANG:
		if val, ok := rhs.(bool); ok {
			f.returnVal = !val
		} else {
			f.returnVal = nil
		}

	case token.TILDE:
		if val, ok := rhs.(int64); !ok {
			f.returnVal = ^val
		} else {
			f.returnVal = nil
		}
	}
	return n, nil
}

func (f *Evaluator) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	lhs := f.returnVal
	n.Rhs.Accept(f, x)
	rhs := f.returnVal

	lv, lok := lhs.(int64)
	rv, rok := rhs.(int64)
	if lok && rok {
		if (n.Op.Type == token.SLASH || n.Op.Type == token.PERCENT) && rv == 0 {
			f.errors.Report(diagnostics.ErrDivByZero, "", n.Op.Line, n.Op.Column, n.Op.Lit)
			panic(&runtimeError{"divided by zero"})
		}
		switch n.Op.Type {
		case "+":
			f.returnVal = lv + rv
			f.returnType = types.SysInt64
		case "-":
			f.returnVal = lv - rv
			f.returnType = types.SysInt64
		case "*":
			f.returnVal = lv * rv
			f.returnType = types.SysInt64
		case "/":
			f.returnVal = lv / rv
			f.returnType = types.SysInt64
		case "%":
			f.returnVal = lv % rv
			f.returnType = types.SysInt64
		case "<<":
			f.returnVal = lv << rv
			f.returnType = types.SysInt64
		case ">>":
			f.returnVal = lv >> rv
			f.returnType = types.SysInt64
		case "&":
			f.returnVal = lv & rv
			f.returnType = types.SysInt64
		case "|":
			f.returnVal = lv | rv
			f.returnType = types.SysInt64
		case ">":
			f.returnVal = lv > rv
			f.returnType = types.SysBool
		case ">=":
			f.returnVal = lv >= rv
			f.returnType = types.SysBool
		case "<":
			f.returnVal = lv < rv
			f.returnType = types.SysBool
		case "<=":
			f.returnVal = lv <= rv
			f.returnType = types.SysBool
		case "!=":
			f.returnVal = lv != rv
			f.returnType = types.SysBool
		case "==":
			f.returnVal = lv == rv
			f.returnType = types.SysBool
		}
	} else if n.Op.Type == token.PLUS {
		f.returnVal = fmt.Sprintf("%v%v", lhs, rhs)
	} else {
		f.returnVal = nil
	}
	return n, nil
}

func (f *Evaluator) VisitCtx(n *ast.CtxExpr, x any) (ast.Node, error) {
	// TODO 构建请求上下文
	return n, nil
}

func (f *Evaluator) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	n.Fn.Accept(f, x)
	fn, ok := f.returnVal.(func([]any) (*types.TypeDef, any, error))
	if !ok {
		astFmt := pretty.NewFormatter()
		name := astFmt.Format(n.Fn)
		f.errors.Report(diagnostics.ErrNotAFunction, "", n.Token.Line, n.Token.Column, name)
		panic(&runtimeError{fmt.Sprintf(diagnostics.ErrNotAFunction.Template, name)})
	}
	args := []any{}
	for _, arg := range n.Args {
		arg.Accept(f, x)
		args = append(args, f.returnVal)
	}
	f.returnType, f.returnVal, f.returnErr = fn(args)
	if f.returnErr != nil {
		// todo 暂时没想到什么情况下会放外面
		// f.fin = false
	}
	f.finished = false
	return n, nil
}

func (f *Evaluator) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	if cond, ok := f.returnVal.(bool); ok {
		if cond {
			n.Then.Accept(f, x)
		} else {
			n.Else.Accept(f, x)
		}
	}
	return n, nil
}

func (f *Evaluator) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	n.Expr.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	var fn funtion
	f.returnVal = fn
	f.returnType = types.SysFunction
	return n, nil
}

func (f *Evaluator) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	// n.Rhs.Accept(f, x)
	// val := f.ret
	// switch left := n.Lhs.(type) {
	// case *ast.IdentExpr:
	// 	f.env.Put(left.Value, val)
	// case *ast.IndexExpr:
	// 	// list[index] = 10
	// 	// tuple[index] = 10
	// 	// map[index] = 10
	// 	left.Lhs.Accept(f, x)
	// 	targetObj := f.ret
	// 	left.Index.Accept(f, x)
	// 	indexObj := f.ret
	// 	switch obj := targetObj.(type) {
	// 	case *oop.ListObj:
	// 		if index, ok := indexObj.(*oop.IntObj); ok {
	// 			obj.Set(int(index.Value), val)
	// 		}
	// 	case *oop.MapObj:
	// 		obj.Put(indexObj, val)
	// 	case *oop.TupleObj:
	// 		if index, ok := indexObj.(*oop.IntObj); ok {
	// 			obj.Set(int(index.Value), val)
	// 		}
	// 	}
	// case *ast.DotExpr:
	// 	left.Lhs.Accept(f, x)
	// 	lhsObj := f.ret
	// 	field := oop.NewString(left.Rhs.Value)
	// 	switch obj := lhsObj.(type) {
	// 	case *oop.MapObj:
	// 		obj.Put(field, val)
	// 	}
	// }
	return n, nil
}

func (f *Evaluator) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	// n.Lhs.Accept(f, x)
	// lv := f.ret
	// n.Index.Accept(f, x)
	// iv := f.ret
	// switch obj := lv.(type) {
	// case *oop.ListObj:
	// 	if index, ok := iv.(*oop.IntObj); ok {
	// 		f.ret = obj.Get(int(index.Value))
	// 	}
	// case *oop.MapObj:
	// 	f.ret = obj.Get(iv)
	// case *oop.TupleObj:
	// 	if index, ok := iv.(*oop.IntObj); ok {
	// 		f.ret = obj.Get(int(index.Value))
	// 	}
	// }
	return n, nil
}

func (f *Evaluator) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	// n.Lhs.Accept(f, x)
	// lhs := f.ret
	// field := oop.NewString(n.Rhs.Value)
	// switch obj := lhs.(type) {
	// case *oop.MapObj:
	// 	f.ret = obj.Get(field)
	// default:
	// 	f.ret = oop.O_NULL
	// }
	return n, nil
}

func (f *Evaluator) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	cond := f.returnVal
	if isTruthy(cond) {
		return n.Then.Accept(f, x)
	}
	if n.Else != nil {
		return n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	n.Value.Accept(f, x)
	_, err := f.symtab.Define(n.Name.Value, sym.KindVariable, f.returnType)
	if err != nil {
		panic(&runtimeError{"变量已存在"})
	}
	return n, nil
}

func (f *Evaluator) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	for {
		n.Cond.Accept(f, x)
		if f.finished || isTruthy(f.returnVal) {
			break
		}
		n.Body.Accept(f, x)
		if f.finished {
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
	f.finished = true
	return n, nil
}

func (f *Evaluator) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	f.symtab.Enter("<block>")
	defer f.symtab.Leave()
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		if f.finished {
			break
		}
	}
	return n, nil
}

func (f *Evaluator) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) { return n, nil }

func NewEvaluator(global *sym.SymbolTable) *Evaluator {
	return &Evaluator{
		symtab: global,
	}
}

type runtimeError struct {
	msg string
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int64:
		return val != 0
	case string:
		return val != ""
	case []any:
		return len(val) > 0
	case map[any]any:
		return len(val) > 0
	case map[string]any:
		return len(val) > 0
	}
	return true
}

func equal(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a == b
}

func inspect(v any) string {
	if v == nil {
		return "nil"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return strconv.FormatInt(val, 10)
	case string:
		return val
	case []any:
		ps := make([]string, len(val))
		for i, e := range val {
			ps[i] = inspect(e)
		}
		return "[" + strings.Join(ps, ",") + "]"
	case map[any]any:
		s := "{"
		first := true
		for k, vv := range val {
			if !first {
				s += ", "
			}
			first = false
			s += inspect(k) + ": " + inspect(vv)
		}
		return s + "}"
	case map[string]any:
		s := "{"
		first := true
		for k, vv := range val {
			if k == "__className" {
				continue
			}
			if !first {
				s += ", "
			}
			first = false
			s += inspect(k) + ": " + inspect(vv)
		}
		return s + "}"
	case funtion:
		return "<fn>"
	default:
		return fmt.Sprint(v)
	}
}
