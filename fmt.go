package mk

import (
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

func (f *PrettyFormatter) Format(p Node) string {
	f.sb.Reset()
	f.level = 0
	p.Accept(f, nil)
	return f.sb.String()
}

func (f *PrettyFormatter) VisitProgram(n *Program, x any) (Node, error) {
	for _, stmt := range n.Statements {
		stmt.Accept(f, x)
	}
	return n, nil
}

// Expr

func (f *PrettyFormatter) VisitIndex(n *IndexExpr, x any) (Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString("[")
	n.Index.Accept(f, x)
	f.sb.WriteString("]")
	return n, nil
}

func (f *PrettyFormatter) VisitDot(n *DotExpr, x any) (Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(".")
	n.Rhs.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitString(n *StringLitExpr, x any) (Node, error) {
	f.sb.WriteString("\"")
	f.sb.WriteString(n.Value)
	f.sb.WriteString("\"")
	return n, nil
}

func (f *PrettyFormatter) VisitInt(n *IntLitExpr, x any) (Node, error) {
	f.sb.WriteString(strconv.Itoa(int(n.Value)))
	return n, nil
}

func (f *PrettyFormatter) VisitBool(n *BoolLitExpr, x any) (Node, error) {
	if n.Value {
		f.sb.WriteString("true")
	} else {
		f.sb.WriteString("false")
	}
	return n, nil
}

func (f *PrettyFormatter) VisitList(n *ListLitExpr, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitTuple(n *TupleLitExpr, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitMap(n *MapLitExpr, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitIdent(n *IdentExpr, x any) (Node, error) {
	f.sb.WriteString(n.Value)
	return n, nil
}

func (f *PrettyFormatter) VisitUnary(n *UnaryExpr, x any) (Node, error) {
	f.sb.WriteString(n.Op.Lit)
	n.Right.Accept(f, nil)
	return n, nil
}

func (f *PrettyFormatter) VisitBinary(n *BinaryExpr, x any) (Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(" ")
	f.sb.WriteString(n.Op.Lit)
	f.sb.WriteString(" ")
	n.Rhs.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitCall(n *CallExpr, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitTernary(n *TernaryExpr, x any) (Node, error) {
	// a ? x : y
	n.Cond.Accept(f, x)
	f.sb.WriteString(" ? ")
	n.Then.Accept(f, x)
	f.sb.WriteString(" : ")
	n.Else.Accept(f, x)
	return n, nil
}

func (f *PrettyFormatter) VisitParen(n *ParenExpr, x any) (Node, error) {
	// ( expr )
	f.sb.WriteString("(")
	n.Expr.Accept(f, x)
	f.sb.WriteString(")")
	return n, nil
}

func (f *PrettyFormatter) VisitFn(n *FnExpr, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitAssign(n *AssignExpr, x any) (Node, error) {
	n.Lhs.Accept(f, x)
	f.sb.WriteString(" = ")
	n.Rhs.Accept(f, x)
	return n, nil
}

// Stmt
func (f *PrettyFormatter) VisitIf(n *IfStmt, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitLet(n *LetStmt, x any) (Node, error) {
	f.writeIndent()
	f.sb.WriteString("let ")
	n.Name.Accept(f, x)
	f.sb.WriteString(" = ")
	n.Value.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitWhile(n *WhileStmt, x any) (Node, error) {
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

func (f *PrettyFormatter) VisitExpr(n *ExprStmt, x any) (Node, error) {
	f.writeIndent()
	n.Expr.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitReturn(n *ReturnStmt, x any) (Node, error) {
	f.writeIndent()
	f.sb.WriteString("return ")
	n.Value.Accept(f, x)
	f.sb.WriteString(";")
	return n, nil
}

func (f *PrettyFormatter) VisitBlock(n *BlockStmt, x any) (Node, error) {
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
func (f *PrettyFormatter) VisitClass(n *ClassStmt, x any) (Node, error) { return n, nil }

func (f *PrettyFormatter) writeIndent() {
	f.sb.WriteString(strings.Repeat(f.ident, f.level))
}
