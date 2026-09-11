package mk

import "strings"

type Visitor interface {
	VisitProgram(*Program, any) (Node, error)

	// Expr
	VisitIdent(*IdentExpr, any) (Node, error)
	VisitLiteral(*LiteralExpr, any) (Node, error)
	VisitUnary(*UnaryExpr, any) (Node, error)
	VisitBinary(*BinaryExpr, any) (Node, error)
	VisitCall(*CallExpr, any) (Node, error)
	VisitTernary(*TernaryExpr, any) (Node, error)
	VisitParen(*ParenExpr, any) (Node, error)
	VisitFn(*FnExpr, any) (Node, error)
	VisitAssign(*AssignExpr, any) (Node, error)

	// Stmt
	VisitIf(*IfStmt, any) (Node, error)
	VisitLet(*LetStmt, any) (Node, error)
	VisitWhile(*WhileStmt, any) (Node, error)
	VisitExpr(*ExprStmt, any) (Node, error)
	VisitReturn(*ReturnStmt, any) (Node, error)
	VisitBlock(*BlockStmt, any) (Node, error)

	// Class
	VisitClass(*ClassStmt, any) (Node, error)
}

type Node interface {
	Text() string
	Accept(Visitor, any) (Node, error)
}

type Statement interface {
	Node
	stmtNode()
}

type Expression interface {
	Node
	exprNode()
}

type Program struct {
	Statements []Statement
}

type IdentExpr struct {
	Token Token
	Value string
}

type UnaryExpr struct {
	Token Token
	Right Expression
}

type BinaryExpr struct {
	Lhs Expression
	Op  Token
	Rhs Expression
}

type CallExpr struct {
	Fn   Expression
	Args []Expression
}

type TernaryExpr struct {
	Token Token
	Cond  Expression
	Then  Expression
	Else  Expression
}

type ArrayAccessExpr struct {
	Expr  Expression
	Index Expression
}

type LiteralExpr struct {
	Token Token
	Value string
}

type AssignExpr struct {
	Token Token
	Lhs   Expression
	Rhs   Expression
}

type FnExpr struct {
	Token Token
	Args  []*IdentExpr
	Body  *BlockStmt
}

type ParenExpr struct {
	Token Token
	Expr  Expression
}

type LetStmt struct {
	Token Token
	Name  *IdentExpr
	Value Expression
}

type IfStmt struct {
	Token Token
	Cond  Expression
	Then  *BlockStmt
	Else  *BlockStmt
}

type WhileStmt struct {
	Token Token
	Cond  Expression
	Body  *BlockStmt
}
type BlockStmt struct {
	Token      Token
	Statements []Statement
}
type ReturnStmt struct {
	Token Token
	Value Expression
}
type ExprStmt struct {
	Token Token
	Expr  Expression
}

type ClassStmt struct {
	Token Token
	Name  *IdentExpr
}

func (e *IdentExpr) exprNode() {
}

func (e *IdentExpr) Text() string {
	return e.Token.Lit
}

func (e *IdentExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitIdent(e, x)
}

func (e *LiteralExpr) exprNode() {
}

func (e *LiteralExpr) Text() string {
	return e.Token.Lit
}

func (e *LiteralExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitLiteral(e, x)
}

func (e *UnaryExpr) exprNode() {
}

func (e *UnaryExpr) Text() string {
	return e.Token.Lit + e.Right.Text()
}

func (e *UnaryExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitUnary(e, x)
}

func (e *BinaryExpr) exprNode() {
}

func (e *BinaryExpr) Text() string {
	return e.Lhs.Text() + " " + e.Op.Lit + " " + e.Rhs.Text()
}

func (e *BinaryExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitBinary(e, x)
}

func (e *CallExpr) exprNode() {
}

func (e *CallExpr) Text() string {
	args := []string{}
	if len(e.Args) > 0 {
		for _, expr := range e.Args {
			args = append(args, expr.Text())
		}
	}
	return e.Fn.Text()
}

func (e *CallExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitCall(e, x)
}

func (e *TernaryExpr) exprNode() {
}

func (e *TernaryExpr) Text() string {
	return e.Token.Lit
}

func (e *TernaryExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitTernary(e, x)
}

func (e *AssignExpr) exprNode() {
}

func (e *AssignExpr) Text() string {
	return e.Token.Lit
}

func (e *AssignExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitAssign(e, x)
}
func (e *ParenExpr) exprNode() {
}

func (e *ParenExpr) Text() string {
	return e.Token.Lit
}

func (e *ParenExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitParen(e, x)
}

func (e *FnExpr) exprNode() {
}

func (e *FnExpr) Text() string {
	return e.Token.Lit
}

func (e *FnExpr) Accept(v Visitor, x any) (Node, error) {
	return v.VisitFn(e, x)
}

func (s *Program) stmtNode() {}

func (s *Program) Text() string {
	if len(s.Statements) > 0 {
		arr := make([]string, len(s.Statements))
		for i, stmt := range s.Statements {
			arr[i] = stmt.Text()
		}
		return strings.Join(arr, "\n")
	}
	return ""
}

func (s *Program) Accept(v Visitor, x any) (Node, error) {
	return v.VisitProgram(s, x)
}

func (s *LetStmt) stmtNode() {}

func (s *LetStmt) Text() string {
	return "let"
}

func (s *LetStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitLet(s, x)
}

func (s *IfStmt) stmtNode() {}

func (s *IfStmt) Text() string {
	return "if"
}

func (s *IfStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitIf(s, x)
}

func (s *WhileStmt) stmtNode() {}

func (s *WhileStmt) Text() string {
	return "while"
}

func (s *WhileStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitWhile(s, x)
}

func (s *BlockStmt) stmtNode() {}

func (s *BlockStmt) Text() string {
	return "let"
}

func (s *BlockStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitBlock(s, x)
}

func (s *ReturnStmt) stmtNode() {}

func (s *ReturnStmt) Text() string {
	return "let"
}

func (s *ReturnStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitReturn(s, x)
}

func (s *ExprStmt) stmtNode() {}

func (s *ExprStmt) Text() string {
	return ""
}

func (s *ExprStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitExpr(s, x)
}

func (s *ClassStmt) stmtNode() {}

func (s *ClassStmt) Text() string {
	return ""
}

func (s *ClassStmt) Accept(v Visitor, x any) (Node, error) {
	return v.VisitClass(s, x)
}
