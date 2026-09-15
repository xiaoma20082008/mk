package pretty

import (
	"mk/internal/ast"
	"strconv"
	"strings"
)

type PrettyFormatter struct {
	sb strings.Builder

	level int
	ident string
}

func NewFormatter() *PrettyFormatter {
	f := &PrettyFormatter{
		level: 0,
		ident: "  ",
	}
	return f
}

func (f *PrettyFormatter) Format(p ast.Node) string {
	f.sb.Reset()
	f.level = 0
	p.Accept(f, nil)
	return f.sb.String()
}

func (f *PrettyFormatter) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
	}
	return n, nil
}

// Expr

func (f *PrettyFormatter) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString("[")
	n.Index.Accept(f, x)
	f.sb.WriteString("]")
	return n, nil
}

func (f *PrettyFormatter) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(".")
	n.Rhs.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	f.sb.WriteString("\"")
	f.sb.WriteString(n.Value)
	f.sb.WriteString("\"")
	return n, nil
}

func (f *PrettyFormatter) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	f.sb.WriteString(strconv.Itoa(int(n.Value)))
	return n, nil
}

func (f *PrettyFormatter) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	if n.Value {
		f.sb.WriteString("true")
	} else {
		f.sb.WriteString("false")
	}
	return n, nil
}

func (f *PrettyFormatter) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	f.sb.WriteString("[")
	for i, v := range n.Value {
		if i > 0 {
			f.sb.WriteString(", ")
		}
		v.Accept(f, x)
	}
	f.sb.WriteString("]")
	return n, nil
}

func (f *PrettyFormatter) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	// (x, y, y,)
	f.sb.WriteString("(")
	for i, v := range n.Value {
		if i > 0 {
			f.sb.WriteString(" ")
		}
		v.Accept(f, x)
		f.sb.WriteString(",")
	}
	f.sb.WriteString(")")
	return n, nil
}

func (f *PrettyFormatter) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	f.sb.WriteString("{")
	i := 0
	sz := len(n.Value)
	for k, v := range n.Value {
		k.Accept(f, x)
		f.sb.WriteString(": ")
		v.Accept(f, x)
		if i < sz-1 {
			f.sb.WriteString(", ")
		}
		i++
	}
	f.sb.WriteString("}")
	return n, nil
}

func (f *PrettyFormatter) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	f.sb.WriteString(n.Value)
	return n, nil
}

func (f *PrettyFormatter) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	f.sb.WriteString(n.Op.Lit)
	n.Right.Accept(f, nil)
	return n, nil
}

func (f *PrettyFormatter) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(" ")
	f.sb.WriteString(n.Op.Lit)
	f.sb.WriteString(" ")
	n.Rhs.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	// add(a, b)
	n.Fn.Accept(f, x)
	f.sb.WriteString("(")
	if len(n.Args) > 0 {
		for i, expr := range n.Args {
			if i > 0 {
				f.sb.WriteString(", ")
			}
			expr.Accept(f, x)
		}
	}
	f.sb.WriteString(")")
	return n, nil
}

func (f *PrettyFormatter) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	// a ? x : y
	n.Cond.Accept(f, x)
	f.sb.WriteString(" ? ")
	n.Then.Accept(f, x)
	f.sb.WriteString(" : ")
	n.Else.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	// ( expr )
	f.sb.WriteString("(")
	n.Expr.Accept(f, x)
	f.sb.WriteString(")")
	return n, nil
}

func (f *PrettyFormatter) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	// fn(x, y) {}
	f.sb.WriteString("fn(")
	for i, arg := range n.Args {
		if i > 0 {
			f.sb.WriteString(", ")
		}
		arg.Accept(f, x)
	}
	f.sb.WriteString(") ")
	n.Body.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(" = ")
	n.Rhs.Accept(f, x)
	return n, nil
}

// Stmt
func (f *PrettyFormatter) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	f.writeIndent()
	f.sb.WriteString("if (")
	n.Cond.Accept(f, x)
	f.sb.WriteString(")")

	n.Then.Accept(f, x)

	if n.Else != nil {
		f.sb.WriteString(" else ")
		n.Else.Accept(f, x)
	}
	return n, nil
}

func (f *PrettyFormatter) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	f.writeIndent()
	f.sb.WriteString("let ")
	n.Name.Accept(f, x)
	f.sb.WriteString(" = ")
	n.Value.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	/*
		while (a != 10) {
		}
	*/
	f.writeIndent()
	f.sb.WriteString("while (")
	n.Cond.Accept(f, x)
	f.sb.WriteString(") ")
	n.Body.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error) {
	f.writeIndent()
	n.Expr.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) {
	f.writeIndent()
	f.sb.WriteString("return ")
	n.Value.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	f.sb.WriteString("{\n")
	f.level++
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
		f.sb.WriteString("\n")
	}
	f.level--
	f.writeIndent()
	f.sb.WriteString("}")
	return n, nil
}

// Class
func (f *PrettyFormatter) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) { return n, nil }

func (f *PrettyFormatter) writeIndent() {
	f.sb.WriteString(strings.Repeat(f.ident, f.level))
}
