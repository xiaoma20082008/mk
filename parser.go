package mk

import (
	"strconv"
)

type Parser interface {
	ParseCode() *Program
	ParseExpr() Expression
	ParseStmt() Statement
}

type prefixParseFn func() Expression
type infixParseFn func(Expression) Expression

type parserImpl struct {
	Parser
	l Lexer
	r DiagnosticReporter

	token     Token
	prevToken Token

	prefixFns map[TokenType]prefixParseFn
	infixFns  map[TokenType]infixParseFn
}

func NewParser(l Lexer, r DiagnosticReporter) *parserImpl {
	p := &parserImpl{
		l: l,
		r: r,

		prefixFns: map[TokenType]prefixParseFn{},
		infixFns:  map[TokenType]infixParseFn{},
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
	p.prefixFns[LPAREN] = p.parseGroupOrTuple
	p.prefixFns[LBRACE] = p.parseMap
	p.prefixFns[LBRACKET] = p.parseList
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
	p.infixFns[QUESTION] = p.parseTernary
	// 4. (
	p.infixFns[LPAREN] = p.parseCall
	// 5. .
	p.infixFns[DOT] = p.parseDot
	// 6. [
	p.infixFns[LBRACKET] = p.parseIndex
	p.nextToken()
	return p
}

func (p *parserImpl) ParseExpr() Expression {
	return p.parseExpr(0)
}

func (p *parserImpl) ParseStmt() Statement {
	if p.token.Type == EOF {
		return nil
	}
	if p.token.Type == SEMI {
		p.nextToken()
	}
	if p.token.Type == EOF {
		return nil
	}
	switch p.token.Type {
	case LET:
		return p.parseLet()
	case IF:
		return p.parseIf()
	case RETURN:
		return p.parseReturn()
	case WHILE:
		return p.parseWhile()
	case LBRACE:
		return p.parseBlock()
	default:
		expr := p.ParseExpr()
		p.expect(SEMI)
		return &ExprStmt{Expr: expr}
	}
}

func (p *parserImpl) ParseCode() *Program {
	program := &Program{
		Statements: []Statement{},
	}
	for p.token.Type != EOF {
		stmt := p.safeParseStmt()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

func (p *parserImpl) safeParseStmt() (stmt Statement) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(parsingError); ok {
				p.crashRecovery()
				stmt = nil
				return
			} else {
				panic(r)
			}
		}
	}()
	return p.ParseStmt()
}

func (p *parserImpl) crashRecovery() {
	p.nextToken()
	for p.token.Type != EOF {
		if p.token.Type == SEMI {
			p.nextToken()
			return
		}
		switch p.token.Type {
		case LET, IF, WHILE, RETURN, FUNCTION:
			return
		}
		p.nextToken()
	}
}

func (p *parserImpl) parseExpr(prec int) Expression {
	var expr Expression
	prefixFn, ok := p.prefixFns[p.token.Type]
	if !ok {
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(ErrNoPrefixParseFunc, line, col, p.token.Lit)
		panic(parsingError{})
	}
	expr = prefixFn()
	for prec < precedence(p.token.Type) {
		infixFn, ok := p.infixFns[p.token.Type]
		if !ok {
			return expr
		}
		expr = infixFn(expr)
	}
	return expr
}

func (p *parserImpl) parsePrefix0() Expression {
	// 1. 字面量: ident
	// 2. 数据类型: string,int,true,false,null,self
	// 3. 一元运算符: +5,-10,!false,~x,typeof(),sizeof()
	// 4. 复合表达式: (), fn, new
	var expr Expression
	switch p.token.Type {
	case IDENT:
		expr = p.parseIdent()
	// bool,int,string,list,map,tuple
	case TRUE, FALSE:
		v, _ := strconv.ParseBool(p.token.Lit)
		expr = &BoolLitExpr{Token: p.token, Value: v}
		p.nextToken()
	case INT:
		v, _ := strconv.ParseInt(p.token.Lit, 10, 64)
		expr = &IntLitExpr{Token: p.token, Value: v}
		p.nextToken()
	case STRING:
		expr = &StringLitExpr{Token: p.token, Value: p.token.Lit}
		p.nextToken()
	case LBRACKET:
		expr = p.parseList()
	case LBRACE:
		expr = p.parseMap()
	case PLUS, MINUS, BANG, TILDE:
		expr = p.parseUnary()
	case LPAREN:
		expr = p.parseGroupOrTuple()
	case FUNCTION:
		expr = p.parseFunction()
	case NEW:
		expr = p.parseNew()
	default:
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(ErrUnknownPrefixKind, line, col, p.token.Lit)
		panic(parsingError{})
	}
	return expr
}

func (p *parserImpl) parseUnary() Expression {
	tok := p.token
	p.nextToken()
	expr := p.parseExpr(precedence(tok.Type))
	return &UnaryExpr{Token: tok, Right: expr}
}

func (p *parserImpl) parseList() Expression {
	// [v1,v2,v3]
	tok := p.token
	p.expect(LBRACKET)
	v := []Expression{}
	if p.token.Type != RBRACKET {
		v = append(v, p.ParseExpr())
		for p.token.Type == COMMA {
			p.nextToken()
			v = append(v, p.ParseExpr())
		}
	}
	p.expect(RBRACKET)
	return &ListLitExpr{Token: tok, Value: v}
}

func (p *parserImpl) parseMap() Expression {
	// {k:v,k:v,}
	tok := p.token
	p.expect(LBRACE)
	v := map[Expression]Expression{}
	if p.token.Type != RBRACE {
		key := p.ParseExpr()
		p.expect(COLON)
		val := p.ParseExpr()
		v[key] = val

		for p.token.Type == COMMA {
			p.nextToken()
			if p.token.Type == RBRACE { // 支持尾随逗号 {a: 1, b: 2, }
				break
			}
			key := p.ParseExpr()
			p.expect(COLON)
			val := p.ParseExpr()
			v[key] = val
		}
	}
	p.expect(RBRACE)
	return &MapLitExpr{Token: tok, Value: v}
}

func (p *parserImpl) parseGroupOrTuple() Expression {
	// Paren: ( x )
	// Tuple: ( x, y,)
	tok := p.token
	p.expect(LPAREN)
	// 处理空元组 ()
	if p.token.Type == RPAREN {
		p.expect(RPAREN)
		return &TupleLitExpr{Token: tok, Value: []Expression{}}
	}
	expr := p.ParseExpr()
	if p.token.Type == COMMA {
		p.nextToken()
		exprs := []Expression{expr}
		for p.token.Type != RPAREN {
			if p.token.Type == RPAREN { // 支持尾随逗号 (1, 2, )
				break
			}
			exprs = append(exprs, p.ParseExpr())
			if p.token.Type == COMMA {
				p.nextToken()
			} else if p.token.Type != RPAREN {
				p.expect(COMMA) // 既没有逗号也没有右括号，强制报缺逗号错误
			}
		}
		p.expect(RPAREN)
		return &TupleLitExpr{Token: tok, Value: exprs}
	}
	p.expect(RPAREN)
	return &ParenExpr{Token: tok, Expr: expr}
}

func (p *parserImpl) parseFunction() Expression {
	// fn(x,y) {}
	tok := p.token
	p.expect(FUNCTION)
	p.expect(LPAREN)
	args := []*IdentExpr{}
	if p.token.Type == IDENT {
		args = append(args, p.parseIdent())
		for p.token.Type == COMMA {
			p.nextToken()
			args = append(args, p.parseIdent())
		}
	}
	p.expect(RPAREN)
	body := p.parseBlock()
	return &FnExpr{Token: tok, Args: args, Body: body}
}

func (p *parserImpl) parseNew() Expression {
	return nil
}

func (p *parserImpl) parseIdent() *IdentExpr {
	if p.token.Type != IDENT {
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(ErrExpectedToken, line, col, p.token.Lit)
		panic(parsingError{})
	}
	e := &IdentExpr{Token: p.token, Value: p.token.Lit}
	p.nextToken()
	return e
}

func (p *parserImpl) parseCall(left Expression) Expression {
	expr := &CallExpr{Fn: left}
	p.expect(LPAREN)

	args := []Expression{}
	if p.token.Type != RPAREN {
		args = append(args, p.parseExpr(0))
		for p.token.Type == COMMA {
			p.nextToken()
			args = append(args, p.parseExpr(0))
		}
	}
	p.expect(RPAREN)
	expr.Args = args
	return expr
}

func (p *parserImpl) parseTernary(left Expression) Expression {
	tok := p.token
	prec := precedence(tok.Type)
	p.nextToken()

	thn := p.parseExpr(0) // 三元中部表达式通常允许从最低优先级重新解析
	p.expect(COLON)
	els := p.parseExpr(prec)

	return &TernaryExpr{Token: tok, Cond: left, Then: thn, Else: els}
}

func (p *parserImpl) parseDot(left Expression) Expression {
	tok := p.token
	p.expect(DOT)
	rhs := p.parseIdent()
	return &DotExpr{Token: tok, Lhs: left, Rhs: rhs}
}

func (p *parserImpl) parseIndex(left Expression) Expression {
	tok := p.token
	p.expect(LBRACKET)
	idx := p.parseExpr(0)
	p.expect(RBRACKET)
	return &IndexExpr{Token: tok, Lhs: left, Index: idx}
}

func (p *parserImpl) parseBinary(left Expression) Expression {
	tok := p.token
	prec := precedence(p.token.Type)
	p.nextToken()
	right := p.parseExpr(prec)
	if tok.Type == ASSIGN {
		return &AssignExpr{Token: tok, Lhs: left, Rhs: right}
	}
	return &BinaryExpr{Op: tok, Lhs: left, Rhs: right}
}

func (p *parserImpl) nextToken() {
	p.prevToken = p.token
	p.token = p.l.NextToken()
}

func (p *parserImpl) peekToken() Token {
	return p.l.Lookahead(1)
}

func (p *parserImpl) parseLet() *LetStmt {
	// let xxx = xxx ;
	p.expect(LET)
	name := p.parseIdent()
	p.expect(ASSIGN)
	expr := p.ParseExpr()
	p.expect(SEMI)
	return &LetStmt{Name: name, Value: expr}
}

func (p *parserImpl) parseIf() *IfStmt {
	tok := p.token
	p.expect(IF)
	p.expect(LPAREN)
	cond := p.ParseExpr()
	p.expect(RPAREN)
	then := p.parseBlock()

	s := &IfStmt{Token: tok, Cond: cond, Then: then}
	if p.token.Type == ELSE {
		p.nextToken()
		s.Else = p.parseBlock()
	}
	return s
}

func (p *parserImpl) parseWhile() *WhileStmt {
	tok := p.token
	p.expect(WHILE)
	p.expect(LPAREN)
	cond := p.ParseExpr()
	p.expect(RPAREN)
	body := p.parseBlock()
	return &WhileStmt{Token: tok, Cond: cond, Body: body}
}

func (p *parserImpl) parseBlock() *BlockStmt {
	tok := p.token
	p.expect(LBRACE)
	stmts := []Statement{}
	for p.token.Type != RBRACE && p.token.Type != EOF {
		stmt := p.safeParseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	p.expect(RBRACE)
	return &BlockStmt{Token: tok, Statements: stmts}
}

func (p *parserImpl) parseReturn() *ReturnStmt {
	// return;
	// return xxx;
	tok := p.token
	p.expect(RETURN)
	var expr Expression
	if p.token.Type != SEMI {
		expr = p.ParseExpr()
	}
	p.expect(SEMI)
	return &ReturnStmt{Token: tok, Value: expr}
}

func (p *parserImpl) expect(kind TokenType) {
	if p.token.Type == kind {
		p.nextToken()
		return
	}

	pos := p.token.Offset
	if p.token.Type == EOF {
		pos = p.prevToken.Offset + len(p.prevToken.Lit)
	}
	line, col := p.l.LineMap(pos)

	p.r.Report(ErrExpectedToken, line, col, string(kind), p.token.Lit)
	panic(parsingError{})
}

type parsingError struct{}
