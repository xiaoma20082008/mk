package mk

import (
	"errors"
	"fmt"
)

type Parser interface {
	ParseCode() (*Program, error)
	ParseExpr() (Expression, error)
	ParseStmt() (Statement, error)
	Errors() []error
}

type prefixParseFn func() (Expression, error)
type infixParseFn func(Expression, int) (Expression, error)

type parserImpl struct {
	Parser
	l *Lexer

	token Token

	prefixFns map[TokenType]prefixParseFn
	infixFns  map[TokenType]infixParseFn

	errors []error
}

func NewParser(l *Lexer) *parserImpl {
	p := &parserImpl{
		l: l,

		prefixFns: map[TokenType]prefixParseFn{},
		infixFns:  map[TokenType]infixParseFn{},

		errors: []error{},
	}
	// 1.
	p.prefixFns[IDENT] = p.parsePrefix0
	// 2.
	p.prefixFns[INT] = p.parsePrefix0
	p.prefixFns[STRING] = p.parsePrefix0
	// 3.
	p.prefixFns[TRUE] = p.parsePrefix0
	p.prefixFns[FALSE] = p.parsePrefix0
	p.prefixFns[NULL] = p.parsePrefix0
	p.prefixFns[SELF] = p.parsePrefix0
	// 4.
	p.prefixFns[PLUS] = p.parseUnary
	p.prefixFns[MINUS] = p.parseUnary
	p.prefixFns[BANG] = p.parseUnary
	p.prefixFns[TILDE] = p.parseUnary
	// 5.
	p.prefixFns[LPAREN] = p.parseGroup
	p.prefixFns[FUNCTION] = p.parseFunction
	p.prefixFns[NEW] = p.parseNew

	// 1. =
	p.infixFns[ASSIGN] = p.parseBinary
	// 2. + - * / % > >= == < <= != << >>
	p.infixFns[PLUS] = p.parseBinary
	p.infixFns[MINUS] = p.parseBinary
	p.infixFns[STAR] = p.parseBinary
	p.infixFns[SLASH] = p.parseBinary
	p.infixFns[PERCENT] = p.parseBinary
	p.infixFns[LT] = p.parseBinary
	p.infixFns[LE] = p.parseBinary
	p.infixFns[EQ] = p.parseBinary
	p.infixFns[GT] = p.parseBinary
	p.infixFns[GE] = p.parseBinary
	p.infixFns[NE] = p.parseBinary
	p.infixFns[LTLT] = p.parseBinary
	p.infixFns[GTGT] = p.parseBinary
	// 3. a? b: c
	p.infixFns[QUESTION] = p.parseBinary
	// 4. (
	p.infixFns[LPAREN] = p.parseBinary
	p.nextToken()
	return p
}

func (p *parserImpl) ParseExpr() (Expression, error) {
	return p.parseExpr(0)
}

func (p *parserImpl) ParseStmt() (Statement, error) {
	switch p.token.Type {
	case LET:
		return p.parseLet()
	case IF:
		return p.parseIf()
	case RETURN:
		return p.parseReturn()
	case FOR:
		return p.parseFor()
	case LBRACE:
		return p.parseBlock()
	default:
		expr, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		// if p.token.Type == SEMI {
		// 	p.nextToken()
		// }
		return &ExprStmt{Expr: expr}, nil
	}
}

func (p *parserImpl) ParseCode() *Program {
	return p.parseProgram()
}

func (p *parserImpl) Errors() []error {
	return p.errors
}

func (p *parserImpl) parseProgram() *Program {
	program := &Program{
		Statements: []Statement{},
	}
	for p.token.Type != EOF {
		stmt, err := p.ParseStmt()
		if err != nil {
			p.errors = append(p.errors, err)
			continue
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

func (p *parserImpl) parseExpr(prec int) (Expression, error) {
	var expr Expression
	var err error
	prefixFn, ok := p.prefixFns[p.token.Type]
	if !ok {
		return nil, errors.New("Prefix parse Function not found for: " + p.token.Lit)
	}
	expr, err = prefixFn()
	if err != nil {
		return nil, err
	}
	for prec < precedence(p.token.Type) {
		infixFn, ok := p.infixFns[p.token.Type]
		if !ok {
			return nil, errors.New("Infix parse Function not found for: " + p.token.Lit)
		}
		expr, err = infixFn(expr, precedence(p.token.Type))
		if err != nil {
			return nil, err
		}
	}
	return expr, nil
}

func (p *parserImpl) parsePrefix0() (Expression, error) {
	// 1. 字面量: ident
	// 2. 数据类型: string,int,true,false,null,self
	// 3. 一元运算符: +5,-10,!false,~x,typeof(),sizeof()
	// 4. 复合表达式: (), fn, new
	var expr Expression
	var err error
	switch p.token.Type {
	case IDENT:
		expr, err = p.parseIdent()
	case INT, STRING:
		expr = &LiteralExpr{Token: p.token, Value: p.token.Lit}
		p.nextToken()
	case TRUE, FALSE, NULL, SELF:
		expr = &LiteralExpr{Token: p.token, Value: p.token.Lit}
		p.nextToken()
	case PLUS, MINUS, BANG, TILDE:
		expr, err = p.parseUnary()
	case LPAREN:
		expr, err = p.parseGroup()
	case FUNCTION:
		expr, err = p.parseFunction()
	case NEW:
		expr, err = p.parseNew()
	default:
		err = errors.New("")
	}
	return expr, err
}

func (p *parserImpl) parseUnary() (Expression, error) {
	tok := p.token
	p.nextToken()
	expr, err := p.parseExpr(precedence(tok.Type))
	if err != nil {
		return nil, err
	}
	return &UnaryExpr{Token: tok, Right: expr}, nil
}

func (p *parserImpl) parseGroup() (Expression, error) {
	tok := p.token
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	p.accept(RPAREN)
	return &ParenExpr{Token: tok, Expr: expr}, nil
}

func (p *parserImpl) parseFunction() (Expression, error) {
	tok := p.token
	p.accept(FUNCTION)
	args := []*IdentExpr{}
	p.accept(LPAREN)
	if p.token.Type == IDENT {
		ident, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		args = append(args, ident)
		for {
			if p.token.Type == COMMA {
				p.nextToken()
				ident, err := p.parseIdent()
				if err != nil {
					return nil, err
				}
				args = append(args, ident)
			} else {
				break
			}
		}
	}
	p.accept(RPAREN)
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &FnExpr{Token: tok, Args: args, Body: body}, nil
}

func (p *parserImpl) parseNew() (Expression, error) {
	return nil, nil
}

func (p *parserImpl) parseIdent() (*IdentExpr, error) {
	if p.token.Type != IDENT {
		return nil, errors.New("")
	}
	e := &IdentExpr{Token: p.token, Value: p.token.Lit}
	p.nextToken()
	return e, nil
}

func (p *parserImpl) parseBinary(expr Expression, precedence int) (Expression, error) {
	switch p.token.Type {
	case ASSIGN:
		tok := p.token
		p.nextToken()
		right, err := p.parseExpr(precedence)
		if err != nil {
			return nil, err
		}
		return &AssignExpr{Token: tok, Lhs: expr, Rhs: right}, nil
	case PLUS, MINUS, STAR, SLASH, LT, LE, GT, GE, NE, EQ, LTLT, GTGT:
		tok := p.token
		p.nextToken()
		right, err := p.parseExpr(precedence)
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Lhs: expr, Op: tok, Rhs: right}, nil
	case LPAREN:
		expr := &CallExpr{Fn: expr}
		p.nextToken()
		if p.token.Type == RPAREN {
			p.nextToken()
			return expr, nil
		}
		args := []Expression{}
		arg0, err := p.parseExpr(0)
		if err != nil {
			return nil, err
		}
		args = append(args, arg0)
		for p.token.Type == COMMA {
			p.nextToken()
			argn, err := p.parseExpr(0)
			if err != nil {
				return nil, err
			}
			args = append(args, argn)
		}
		p.accept(RPAREN)
		expr.Args = args
		if p.token.Type == SEMI {
			p.nextToken()
		}
		return expr, nil
	case QUESTION:
		tok := p.token
		p.nextToken()
		thn, err := p.parseExpr(precedence)
		if err != nil {
			return nil, err
		}
		p.accept(COLON)
		els, err := p.parseExpr(precedence)
		if err != nil {
			return nil, err
		}
		return &TernaryExpr{Token: tok, Cond: expr, Then: thn, Else: els}, nil
	}
	return nil, nil
}

func (p *parserImpl) nextToken() {
	p.token = p.l.NextToken()
}

func (p *parserImpl) peekToken() Token {
	return p.l.Lookhead(1)
}

func (p *parserImpl) parseLet() (*LetStmt, error) {
	if !p.accept(LET) {
		return nil, errors.New("Expect let")
	}
	name := &IdentExpr{Token: p.token, Value: p.token.Lit}
	p.nextToken()

	if !p.accept(ASSIGN) {
		return nil, errors.New("Expect =")
	}
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}

	if p.token.Type == SEMI {
		p.nextToken()
	}
	return &LetStmt{Name: name, Value: expr}, nil
}

func (p *parserImpl) parseIf() (*IfStmt, error) {
	tok := p.token
	if !p.accept(IF) {
		return nil, errors.New("Expect if")
	}
	p.accept(LPAREN)
	cond, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	p.accept(RPAREN)
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	s := &IfStmt{}
	s.Token = tok
	s.Cond = cond
	s.Then = then

	if p.token.Type == ELSE {
		p.nextToken()
		els, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		s.Else = els
	}
	return s, nil
}

func (p *parserImpl) parseFor() (*ForStmt, error) {
	tok := p.token
	p.accept(FOR)
	p.accept(LPAREN)

	var init Statement
	var cond Expression
	var post Expression
	var err error
	if p.token.Type != SEMI {
		init, err = p.ParseStmt()
		if err != nil {
			return nil, err
		}
	}
	p.accept(SEMI)
	if p.token.Type != SEMI {
		cond, err = p.ParseExpr()
		if err != nil {
			return nil, err
		}
	}
	p.accept(SEMI)
	if p.token.Type != RPAREN {
		post, err = p.ParseExpr()
		if err != nil {
			return nil, err
		}
	}
	p.accept(RPAREN)
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ForStmt{Token: tok, Init: init, Cond: cond, Post: post, Body: body}, nil
}

func (p *parserImpl) parseBlock() (*BlockStmt, error) {
	tok := p.token
	p.accept(LBRACE)
	stmts := []Statement{}
	for p.token.Type != RBRACE {
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	p.accept(RBRACE)
	return &BlockStmt{Token: tok, Statements: stmts}, nil
}

func (p *parserImpl) parseReturn() (*ReturnStmt, error) {
	tok := p.token
	p.accept(RETURN)
	expr, err := p.ParseExpr()
	p.accept(SEMI)
	return &ReturnStmt{Token: tok, Value: expr}, err
}

func (p *parserImpl) accept(kind TokenType) bool {
	if p.token.Type == kind {
		p.nextToken()
		return true
	} else {
		p.failed(p.l.pos, fmt.Sprintf("Expect %s, but got %s", kind, p.token.Type))
		return false
	}
}

func (p *parserImpl) failed(pos int, msg string) {}
