package interp

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/pretty"
	"mk/internal/sym"
	"mk/internal/token"
	"mk/internal/types"
)

type Evaluator struct {
	errors diagnostics.DiagnosticReporter
	// 编译期: 类型/作用域信息
	symtab *sym.SymbolTable
	// env 运行期: 环境
	env *Env

	returnType *types.TypeDef
	returnVal  any
	returnErr  any
	finished   bool
}

func (e *Evaluator) EvalWithWriter(p *ast.Program, w io.Writer) any {
	res := e.Eval(p)
	str := fmt.Sprintln(inspect(res))
	w.Write([]byte(str))
	return res
}

func (e *Evaluator) Eval(p *ast.Program) any {
	e.finished = false
	e.returnVal = nil
	e.returnType = nil
	e.returnErr = nil

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

func (f *Evaluator) report(desc diagnostics.DiagnosticDescriptor, line, col int, args ...any) {
	if f.errors != nil {
		f.errors.Report(desc, "", line, col, args...)
	}
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
	// 1. 运行期值环境优先
	if v, ok := f.env.Get(n.Value); ok {
		f.returnVal = v
		if s, ok := f.symtab.Resolve(n.Value); ok {
			f.returnType = s.Type
		}
		return n, nil
	}
	// 2. 内建函数
	if b, ok := builtins[n.Value]; ok {
		f.returnVal = b
		f.returnType = types.SysFunction
		return n, nil
	}
	// 3. 只有类型信息、尚无运行期值的符号
	if s, ok := f.symtab.Resolve(n.Value); ok {
		f.returnVal = nil
		f.returnType = s.Type
		return n, nil
	}

	f.report(diagnostics.ErrUndefinedIdentifier, n.Token.Line, n.Token.Column, n.Token.Lit)
	panic(&runtimeError{fmt.Sprintf(diagnostics.ErrUndefinedIdentifier.Template, n.Token.Lit)})
}

func (f *Evaluator) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	m := map[any]any{}
	for k, v := range n.Value {
		k.Accept(f, x)
		key := f.returnVal
		v.Accept(f, x)
		m[key] = f.returnVal
	}
	f.returnVal = m
	f.returnType = types.SysObject
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
	ret := make([]any, 0, len(n.Value))
	for _, expr := range n.Value {
		expr.Accept(f, x)
		ret = append(ret, f.returnVal)
	}
	f.returnVal = ret
	f.returnType = types.SysObject
	return n, nil
}

func (f *Evaluator) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	vals := make([]any, 0, len(n.Value))
	for _, expr := range n.Value {
		expr.Accept(f, x)
		vals = append(vals, f.returnVal)
	}
	f.returnVal = vals
	f.returnType = types.SysObject
	return n, nil
}

func (f *Evaluator) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	f.returnVal = n.Value
	f.returnType = types.SysBool
	return n, nil
}

func (f *Evaluator) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	n.Right.Accept(f, x)
	rhs := f.returnVal

	switch n.Op.Type {
	case token.PLUS:
		if _, ok := rhs.(int64); ok {
			f.returnType = types.SysInt64
		} else {
			f.returnVal = nil
		}
	case token.MINUS:
		if val, ok := rhs.(int64); ok {
			f.returnVal = -val
			f.returnType = types.SysInt64
		} else {
			f.returnVal = nil
		}
	case token.BANG:
		f.returnVal = !isTruthy(rhs)
		f.returnType = types.SysBool
	case token.TILDE:
		if val, ok := rhs.(int64); ok {
			f.returnVal = ^val
			f.returnType = types.SysInt64
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

	// 相等 / 不等：（容器用 reflect.DeepEqual，避免对 map/slice 用 == 触发 panic）
	if n.Op.Type == token.EQ || n.Op.Type == token.NE {
		eq := valuesEqual(lhs, rhs)
		if n.Op.Type == token.NE {
			eq = !eq
		}
		f.returnVal = eq
		f.returnType = types.SysBool
		return n, nil
	}

	// 字符串
	ls, lok := lhs.(string)
	rs, rok := rhs.(string)
	if lok && rok {
		switch n.Op.Type {
		case token.PLUS:
			f.returnVal = ls + rs
			f.returnType = types.SysString
		case token.LT:
			f.returnVal = ls < rs
			f.returnType = types.SysBool
		case token.LE:
			f.returnVal = ls <= rs
			f.returnType = types.SysBool
		case token.GT:
			f.returnVal = ls > rs
			f.returnType = types.SysBool
		case token.GE:
			f.returnVal = ls >= rs
			f.returnType = types.SysBool
		default:
			f.returnVal = nil
		}
		return n, nil
	}

	lv, lok := lhs.(int64)
	rv, rok := rhs.(int64)
	if !lok || !rok {
		// 一边是字符串的 +：做拼接（与语言的宽松语义保持一致）
		if n.Op.Type == token.PLUS {
			if _, ok := lhs.(string); ok {
				f.returnVal = inspect(lhs) + inspect(rhs)
				f.returnType = types.SysString
				return n, nil
			}
			if _, ok := rhs.(string); ok {
				f.returnVal = inspect(lhs) + inspect(rhs)
				f.returnType = types.SysString
				return n, nil
			}
		}
		f.returnVal = nil
		return n, nil
	}

	if (n.Op.Type == token.SLASH || n.Op.Type == token.PERCENT) && rv == 0 {
		f.report(diagnostics.ErrDivByZero, n.Op.Line, n.Op.Column, n.Op.Lit)
		panic(&runtimeError{"divided by zero"})
	}
	if (n.Op.Type == token.LTLT || n.Op.Type == token.GTGT) && rv < 0 {
		panic(&runtimeError{"negative shift amount"})
	}

	switch n.Op.Type {
	case token.PLUS:
		f.returnVal = lv + rv
		f.returnType = types.SysInt64
	case token.MINUS:
		f.returnVal = lv - rv
		f.returnType = types.SysInt64
	case token.STAR:
		f.returnVal = lv * rv
		f.returnType = types.SysInt64
	case token.SLASH:
		f.returnVal = lv / rv
		f.returnType = types.SysInt64
	case token.PERCENT:
		f.returnVal = lv % rv
		f.returnType = types.SysInt64
	case token.LTLT:
		f.returnVal = lv << uint(rv)
		f.returnType = types.SysInt64
	case token.GTGT:
		f.returnVal = lv >> uint(rv)
		f.returnType = types.SysInt64
	case token.BIT_AND:
		f.returnVal = lv & rv
		f.returnType = types.SysInt64
	case token.BIT_OR:
		f.returnVal = lv | rv
		f.returnType = types.SysInt64
	case token.BIT_XOR:
		f.returnVal = lv ^ rv
		f.returnType = types.SysInt64
	case token.GT:
		f.returnVal = lv > rv
		f.returnType = types.SysBool
	case token.GE:
		f.returnVal = lv >= rv
		f.returnType = types.SysBool
	case token.LT:
		f.returnVal = lv < rv
		f.returnType = types.SysBool
	case token.LE:
		f.returnVal = lv <= rv
		f.returnType = types.SysBool
	default:
		f.returnVal = nil
	}
	return n, nil
}

func (f *Evaluator) VisitCtx(n *ast.CtxExpr, x any) (ast.Node, error) {
	// TODO 构建请求上下文；当前先明确返回 null，避免残留上一个表达式的值。
	f.returnVal = nil
	return n, nil
}

func (f *Evaluator) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	n.Fn.Accept(f, x)
	fn, ok := f.returnVal.(funtion)
	if !ok {
		astFmt := pretty.NewFormatter()
		name := astFmt.Format(n.Fn)
		f.report(diagnostics.ErrNotAFunction, n.Token.Line, n.Token.Column, name)
		panic(&runtimeError{fmt.Sprintf(diagnostics.ErrNotAFunction.Template, name)})
	}

	args := make([]any, 0, len(n.Args))
	for _, arg := range n.Args {
		arg.Accept(f, x)
		args = append(args, f.returnVal)
	}

	typ, val, err := fn(args)
	if err != nil {
		panic(&runtimeError{err.Error()})
	}

	//f.callFunction(args)

	f.returnType = typ
	f.returnVal = val
	return n, nil
}

func (f *Evaluator) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	if isTruthy(f.returnVal) {
		n.Then.Accept(f, x)
	} else {
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	n.Expr.Accept(f, x)
	return n, nil
}

func (f *Evaluator) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	captured := f.env
	params := n.Args
	body := n.Body

	f.returnVal = funtion(func(args []any) (*types.TypeDef, any, error) {
		return f.callFunction(params, body, captured, args)
	})
	f.returnType = types.SysFunction
	return n, nil
}

// callFunction 执行一次函数调用：保存调用方状态 -> 新建调用环境绑定参数 -> 执行函数体 -> 恢复调用方状态。
func (f *Evaluator) callFunction(params []*ast.IdentExpr, body *ast.BlockStmt, captured *Env, args []any) (*types.TypeDef, any, error) {
	if len(args) != len(params) {
		return nil, nil, fmt.Errorf("wrong number of arguments: want %d, got %d", len(params), len(args))
	}

	savedEnv := f.env
	savedType, savedVal, savedErr := f.returnType, f.returnVal, f.returnErr
	savedFinished := f.finished

	callEnv := NewEnv(captured)
	f.symtab.Enter("<fn>")
	defer func() {
		f.symtab.Leave()
		f.env = savedEnv
		f.returnType, f.returnVal, f.returnErr = savedType, savedVal, savedErr
		f.finished = savedFinished
	}()

	for i, p := range params {
		callEnv.Define(p.Value, args[i])
		_, _ = f.symtab.Define(p.Value, sym.KindParam, nil)
	}

	f.env = callEnv
	f.finished = false
	f.returnVal = nil
	f.returnType = nil
	f.returnErr = nil

	body.Accept(f, nil)

	resultType := f.returnType
	resultVal := f.returnVal
	var resultErr error
	if f.returnErr != nil {
		if e, ok := f.returnErr.(error); ok {
			resultErr = e
		}
	}
	return resultType, resultVal, resultErr
}

func (f *Evaluator) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	n.Rhs.Accept(f, x)
	val := f.returnVal

	switch lhs := n.Lhs.(type) {
	case *ast.IdentExpr:
		if !f.env.Assign(lhs.Value, val) {
			f.env.Define(lhs.Value, val)
			if _, ok := f.symtab.Resolve(lhs.Value); !ok {
				_, _ = f.symtab.Define(lhs.Value, sym.KindVariable, f.returnType)
			}
		}
	case *ast.IndexExpr:
		lhs.Lhs.Accept(f, x)
		target := f.returnVal
		lhs.Index.Accept(f, x)
		idx := f.returnVal
		f.setIndex(target, idx, val)
	case *ast.DotExpr:
		lhs.Lhs.Accept(f, x)
		target := f.returnVal
		f.setField(target, lhs.Rhs.Value, val)
	default:
		panic(&runtimeError{fmt.Sprintf("unsupported assignment target: %T", n.Lhs)})
	}

	f.returnVal = val
	return n, nil
}

func (f *Evaluator) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	target := f.returnVal
	n.Index.Accept(f, x)
	idx := f.returnVal

	switch obj := target.(type) {
	case []any:
		i, ok := idx.(int64)
		if !ok {
			panic(&runtimeError{"list index must be INT"})
		}
		if i < 0 || int(i) >= len(obj) {
			f.returnVal = nil
			return n, nil
		}
		f.returnVal = obj[i]
	case map[any]any:
		if !isHashable(idx) {
			panic(&runtimeError{fmt.Sprintf("unhashable map key: %T", idx)})
		}
		f.returnVal = obj[idx]
	case map[string]any:
		k, ok := idx.(string)
		if !ok {
			panic(&runtimeError{"map key must be STRING"})
		}
		f.returnVal = obj[k]
	case string:
		i, ok := idx.(int64)
		if !ok {
			panic(&runtimeError{"string index must be INT"})
		}
		runes := []rune(obj)
		if i < 0 || int(i) >= len(runes) {
			f.returnVal = nil
			return n, nil
		}
		f.returnVal = string(runes[i])
	default:
		panic(&runtimeError{fmt.Sprintf("%T is not indexable", target)})
	}
	return n, nil
}

func (f *Evaluator) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	switch obj := f.returnVal.(type) {
	case map[any]any:
		f.returnVal = obj[n.Rhs.Value]
	case map[string]any:
		f.returnVal = obj[n.Rhs.Value]
	default:
		f.returnVal = nil
	}
	return n, nil
}

func (f *Evaluator) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	n.Cond.Accept(f, x)
	if isTruthy(f.returnVal) {
		return n.Then.Accept(f, x)
	}
	if n.Else != nil {
		return n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *Evaluator) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	n.Value.Accept(f, x)
	val := f.returnVal

	if _, err := f.symtab.Define(n.Name.Value, sym.KindVariable, f.returnType); err != nil {
		// 同一作用域内重复 let：REPL 里非常常见（每次重跑同一段脚本都会命中），
		// 这里退化为"覆盖"语义，只更新符号表里的类型，不再直接 panic。
		if s, ok := f.symtab.Resolve(n.Name.Value); ok {
			s.Type = f.returnType
		}
	}
	f.env.Define(n.Name.Value, val)
	return n, nil
}

func (f *Evaluator) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	for {
		n.Cond.Accept(f, x)
		if f.finished || !isTruthy(f.returnVal) {
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
	if n.Value != nil {
		n.Value.Accept(f, x)
	} else {
		f.returnVal = nil
	}
	f.finished = true
	return n, nil
}

func (f *Evaluator) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	f.symtab.Enter("<block>")
	savedEnv := f.env
	f.env = NewEnv(savedEnv)
	defer func() {
		f.env = savedEnv
		f.symtab.Leave()
	}()

	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		if f.finished {
			break
		}
	}
	return n, nil
}

func (f *Evaluator) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) { return n, nil }

func NewEvaluator(global *sym.SymbolTable, reporters ...diagnostics.DiagnosticReporter) *Evaluator {
	e := &Evaluator{
		symtab: global,
		env:    NewEnv(nil),
	}
	if len(reporters) > 0 {
		e.errors = reporters[0]
	}
	return e
}

// ------------------------------------------------------------------------------------------
// 赋值辅助
// ------------------------------------------------------------------------------------------

func (f *Evaluator) setIndex(target, idx, val any) {
	switch obj := target.(type) {
	case []any:
		i, ok := idx.(int64)
		if !ok {
			panic(&runtimeError{"list index must be INT"})
		}
		if i < 0 || int(i) >= len(obj) {
			panic(&runtimeError{fmt.Sprintf("list index out of range: %d", i)})
		}
		obj[i] = val
	case map[any]any:
		if !isHashable(idx) {
			panic(&runtimeError{fmt.Sprintf("unhashable map key: %T", idx)})
		}
		obj[idx] = val
	case map[string]any:
		k, ok := idx.(string)
		if !ok {
			panic(&runtimeError{"map key must be STRING"})
		}
		obj[k] = val
	default:
		panic(&runtimeError{fmt.Sprintf("cannot assign index on %T", target)})
	}
}

func (f *Evaluator) setField(target any, field string, val any) {
	switch obj := target.(type) {
	case map[any]any:
		obj[field] = val
	case map[string]any:
		obj[field] = val
	default:
		panic(&runtimeError{fmt.Sprintf("cannot set field on %T", target)})
	}
}

// ------------------------------------------------------------------------------------------
// 运行期工具
// ------------------------------------------------------------------------------------------

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

// valuesEqual 对任意运行期值做相等比较；容器走 reflect.DeepEqual，避免 == 对不可比较类型 panic。
func valuesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.DeepEqual(a, b)
}

// isHashable 判断值能否作为 Go map 的 key。
func isHashable(v any) bool {
	switch v.(type) {
	case nil, bool, int64, string:
		return true
	default:
		return false
	}
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
