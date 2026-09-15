package parser

import (
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/lexer"
	"mk/internal/token"
	"strconv"
)

type Parser interface {
	ParseCode() *ast.Program
	ParseExpr() ast.Expression
	ParseStmt() ast.Statement
}

type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

type parserImpl struct {
	Parser
	l lexer.Lexer
	r diagnostics.DiagnosticReporter

	token     token.Token
	prevToken token.Token

	prefixFns map[token.TokenType]prefixParseFn
	infixFns  map[token.TokenType]infixParseFn

	depth int
}

func NewParser(l lexer.Lexer, r diagnostics.DiagnosticReporter) *parserImpl {
	p := &parserImpl{
		l: l,
		r: r,

		prefixFns: map[token.TokenType]prefixParseFn{},
		infixFns:  map[token.TokenType]infixParseFn{},
	}

	// parsePrefix主要包括：
	// 1. 字面量，如：10，20，"Tom"
	// 2. 标识符，如：id，name
	// 3. 前缀运算符，如：-，!，~
	// 4. 分组符号，如：(
	// 5. 其它符号，如：[ （数组），{ （对象/Map）

	// parseInfix主要包括：
	// 1. 二元中缀运算符，如：+，-，*，/ 等
	// 2. 后缀运算符，如：!（阶乘），++，--
	// 3. 特殊运算符，如：函数调用add()、索引访问arr[10]、字段访问obj.field

	// 1.
	p.prefixFns[token.IDENT] = p.parsePrefix0
	// 2.
	p.prefixFns[token.INT] = p.parsePrefix0
	p.prefixFns[token.STRING] = p.parsePrefix0
	// 3.
	p.prefixFns[token.TRUE] = p.parsePrefix0
	p.prefixFns[token.FALSE] = p.parsePrefix0
	p.prefixFns[token.NULL] = p.parsePrefix0
	p.prefixFns[token.SELF] = p.parsePrefix0
	// 4.
	p.prefixFns[token.PLUS] = p.parsePrefix0
	p.prefixFns[token.MINUS] = p.parsePrefix0
	p.prefixFns[token.BANG] = p.parsePrefix0
	p.prefixFns[token.TILDE] = p.parsePrefix0
	// 5.
	p.prefixFns[token.LPAREN] = p.parsePrefix0
	p.prefixFns[token.LBRACE] = p.parsePrefix0
	p.prefixFns[token.LBRACKET] = p.parsePrefix0
	p.prefixFns[token.FUNCTION] = p.parsePrefix0
	p.prefixFns[token.NEW] = p.parsePrefix0

	// 1. =
	p.infixFns[token.ASSIGN] = p.parseBinary
	// 2. + - * / % > >= == < <= != << >>
	p.infixFns[token.PLUS] = p.parseBinary
	p.infixFns[token.MINUS] = p.parseBinary
	p.infixFns[token.STAR] = p.parseBinary
	p.infixFns[token.SLASH] = p.parseBinary
	p.infixFns[token.PERCENT] = p.parseBinary
	p.infixFns[token.LT] = p.parseBinary
	p.infixFns[token.LE] = p.parseBinary
	p.infixFns[token.EQ] = p.parseBinary
	p.infixFns[token.GT] = p.parseBinary
	p.infixFns[token.GE] = p.parseBinary
	p.infixFns[token.NE] = p.parseBinary
	p.infixFns[token.LTLT] = p.parseBinary
	p.infixFns[token.GTGT] = p.parseBinary
	// 3. a? b: c
	p.infixFns[token.QUESTION] = p.parseTernary
	// 4. (
	p.infixFns[token.LPAREN] = p.parseCall
	// 5. .
	p.infixFns[token.DOT] = p.parseDot
	// 6. [
	p.infixFns[token.LBRACKET] = p.parseIndex
	p.nextToken()
	return p
}

func (p *parserImpl) ParseExpr() ast.Expression {
	return p.parseExpr(0)
}

func (p *parserImpl) ParseStmt() ast.Statement {
	if p.token.Type == token.EOF {
		return nil
	}
	if p.token.Type == token.SEMI {
		p.nextToken()
	}
	if p.token.Type == token.EOF {
		return nil
	}
	switch p.token.Type {
	case token.LET:
		return p.parseLet()
	case token.IF:
		return p.parseIf()
	case token.RETURN:
		return p.parseReturn()
	case token.WHILE:
		return p.parseWhile()
	case token.LBRACE:
		return p.parseBlock()
	default:
		expr := p.ParseExpr()
		p.expect(token.SEMI)
		return &ast.ExprStmt{Expr: expr}
	}
}

func (p *parserImpl) ParseCode() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}
	for p.token.Type != token.EOF {
		stmt := p.safeParseStmt()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

func (p *parserImpl) safeParseStmt() (stmt ast.Statement) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*parsingError); ok {
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
	for p.token.Type != token.EOF {
		if p.token.Type == token.SEMI {
			p.nextToken()
			return
		}
		switch p.token.Type {
		case token.LET, token.IF, token.WHILE, token.RETURN, token.FUNCTION:
			return
		}
		p.nextToken()
	}
}

func (p *parserImpl) parseExpr(prec int) ast.Expression {
	var expr ast.Expression
	prefixFn, ok := p.prefixFns[p.token.Type]
	if !ok {
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(diagnostics.ErrNoPrefixParseFunc, line, col, p.token.Lit)
		panic(&parsingError{"no prefix parse func found for token: " + p.token.Lit})
	}
	expr = prefixFn()
	for prec < token.GetPrecedence(p.token.Type) {
		infixFn, ok := p.infixFns[p.token.Type]
		if !ok {
			return expr
		}
		expr = infixFn(expr)
	}
	return expr
}

func (p *parserImpl) parsePrefix0() ast.Expression {
	// 1. 字面量: ident
	// 2. 数据类型: string,int,true,false,null,self
	// 3. 一元运算符: +5,-10,!false,~x,typeof(),sizeof()
	// 4. 复合表达式: (), fn, new
	var expr ast.Expression
	switch p.token.Type {
	case token.IDENT:
		expr = p.parseIdent()
	// bool,int,string,list,map,tuple
	case token.TRUE, token.FALSE:
		v, _ := strconv.ParseBool(p.token.Lit)
		expr = &ast.BoolLitExpr{Token: p.token, Value: v}
		p.nextToken()
	case token.INT:
		v, _ := strconv.ParseInt(p.token.Lit, 10, 64)
		expr = &ast.IntLitExpr{Token: p.token, Value: v}
		p.nextToken()
	case token.STRING:
		expr = &ast.StringLitExpr{Token: p.token, Value: p.token.Lit}
		p.nextToken()
	case token.LBRACKET:
		expr = p.parseList()
	case token.LBRACE:
		expr = p.parseMap()
	case token.PLUS, token.MINUS, token.BANG, token.TILDE:
		expr = p.parseUnary()
	case token.LPAREN:
		expr = p.parseGroupOrTuple()
	case token.FUNCTION:
		expr = p.parseFunction()
	case token.NEW:
		expr = p.parseNew()
	default:
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(diagnostics.ErrUnknownPrefixKind, line, col, p.token.Lit)
		panic(&parsingError{"unknown token: " + p.token.Lit})
	}
	return expr
}

func (p *parserImpl) parseUnary() ast.Expression {
	tok := p.token
	p.nextToken()
	expr := p.parseExpr(token.GetPrecedence(tok.Type))
	return &ast.UnaryExpr{Op: tok, Right: expr}
}

func (p *parserImpl) parseList() ast.Expression {
	// [v1,v2,v3]
	tok := p.token
	p.expect(token.LBRACKET)
	v := []ast.Expression{}
	if p.token.Type != token.RBRACKET {
		v = append(v, p.ParseExpr())
		for p.token.Type == token.COMMA {
			p.nextToken()
			v = append(v, p.ParseExpr())
		}
	}
	p.expect(token.RBRACKET)
	return &ast.ListLitExpr{Token: tok, Value: v}
}

func (p *parserImpl) parseMap() ast.Expression {
	// {k:v,k:v,}
	tok := p.token
	p.expect(token.LBRACE)
	v := map[ast.Expression]ast.Expression{}
	if p.token.Type != token.RBRACE {
		key := p.ParseExpr()
		p.expect(token.COLON)
		val := p.ParseExpr()
		v[key] = val

		for p.token.Type == token.COMMA {
			p.nextToken()
			if p.token.Type == token.RBRACE { // 支持尾随逗号 {a: 1, b: 2, }
				break
			}
			key := p.ParseExpr()
			p.expect(token.COLON)
			val := p.ParseExpr()
			v[key] = val
		}
	}
	p.expect(token.RBRACE)
	return &ast.MapLitExpr{Token: tok, Value: v}
}

func (p *parserImpl) parseGroupOrTuple() ast.Expression {
	// Paren: ( x )
	// Tuple: ( x, y,)
	tok := p.token
	p.expect(token.LPAREN)
	// 处理空元组 ()
	if p.token.Type == token.RPAREN {
		p.expect(token.RPAREN)
		return &ast.TupleLitExpr{Token: tok, Value: []ast.Expression{}}
	}
	expr := p.ParseExpr()
	if p.token.Type == token.COMMA {
		p.nextToken()
		exprs := []ast.Expression{expr}
		for p.token.Type != token.RPAREN {
			if p.token.Type == token.RPAREN { // 支持尾随逗号 (1, 2, )
				break
			}
			exprs = append(exprs, p.ParseExpr())
			if p.token.Type == token.COMMA {
				p.nextToken()
			} else if p.token.Type != token.RPAREN {
				p.expect(token.COMMA) // 既没有逗号也没有右括号，强制报缺逗号错误
			}
		}
		p.expect(token.RPAREN)
		return &ast.TupleLitExpr{Token: tok, Value: exprs}
	}
	p.expect(token.RPAREN)
	return &ast.ParenExpr{Token: tok, Expr: expr}
}

func (p *parserImpl) parseFunction() ast.Expression {
	// fn(x,y) {}
	tok := p.token
	p.expect(token.FUNCTION)
	p.expect(token.LPAREN)
	args := []*ast.IdentExpr{}
	if p.token.Type == token.IDENT {
		args = append(args, p.parseIdent())
		for p.token.Type == token.COMMA {
			p.nextToken()
			args = append(args, p.parseIdent())
		}
	}
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.FnExpr{Token: tok, Args: args, Body: body}
}

func (p *parserImpl) parseNew() ast.Expression {
	return nil
}

func (p *parserImpl) parseIdent() *ast.IdentExpr {
	if p.token.Type != token.IDENT {
		line, col := p.l.LineMap(p.token.Offset)
		p.r.Report(diagnostics.ErrExpectedToken, line, col, p.token.Lit)
		panic(&parsingError{"expect an ident, but got: " + p.token.Lit})
	}
	e := &ast.IdentExpr{Token: p.token, Value: p.token.Lit}
	p.nextToken()
	return e
}

func (p *parserImpl) parseCall(left ast.Expression) ast.Expression {
	expr := &ast.CallExpr{Token: p.token, Fn: left}
	p.expect(token.LPAREN)

	args := []ast.Expression{}
	if p.token.Type != token.RPAREN {
		args = append(args, p.parseExpr(0))
		for p.token.Type == token.COMMA {
			p.nextToken()
			args = append(args, p.parseExpr(0))
		}
	}
	p.expect(token.RPAREN)
	expr.Args = args
	return expr
}

func (p *parserImpl) parseTernary(left ast.Expression) ast.Expression {
	tok := p.token
	prec := token.GetPrecedence(tok.Type)
	p.nextToken()

	thn := p.parseExpr(0) // 三元中部表达式通常允许从最低优先级重新解析
	p.expect(token.COLON)
	els := p.parseExpr(prec)

	return &ast.TernaryExpr{Token: tok, Cond: left, Then: thn, Else: els}
}

func (p *parserImpl) parseDot(left ast.Expression) ast.Expression {
	tok := p.token
	p.expect(token.DOT)
	rhs := p.parseIdent()
	return &ast.DotExpr{Token: tok, Lhs: left, Rhs: rhs}
}

func (p *parserImpl) parseIndex(left ast.Expression) ast.Expression {
	tok := p.token
	p.expect(token.LBRACKET)
	idx := p.parseExpr(0)
	p.expect(token.RBRACKET)
	return &ast.IndexExpr{Token: tok, Lhs: left, Index: idx}
}

func (p *parserImpl) parseBinary(left ast.Expression) ast.Expression {
	tok := p.token
	prec := token.GetPrecedence(p.token.Type)
	p.nextToken()
	right := p.parseExpr(prec)
	if tok.Type == token.ASSIGN {
		return &ast.AssignExpr{Token: tok, Lhs: left, Rhs: right}
	}
	return &ast.BinaryExpr{Op: tok, Lhs: left, Rhs: right}
}

func (p *parserImpl) nextToken() {
	p.prevToken = p.token
	p.token = p.l.NextToken()
}

func (p *parserImpl) peekToken() token.Token {
	return p.l.Lookahead(1)
}

func (p *parserImpl) parseLet() *ast.LetStmt {
	// let xxx = xxx ;
	p.expect(token.LET)
	name := p.parseIdent()
	p.expect(token.ASSIGN)
	expr := p.ParseExpr()
	p.expect(token.SEMI)
	return &ast.LetStmt{Name: name, Value: expr}
}

func (p *parserImpl) parseIf() *ast.IfStmt {
	tok := p.token
	p.expect(token.IF)
	p.expect(token.LPAREN)
	cond := p.ParseExpr()
	p.expect(token.RPAREN)
	then := p.parseBlock()

	s := &ast.IfStmt{Token: tok, Cond: cond, Then: then}
	if p.token.Type == token.ELSE {
		p.nextToken()
		s.Else = p.parseBlock()
	}
	return s
}

func (p *parserImpl) parseWhile() *ast.WhileStmt {
	tok := p.token
	p.expect(token.WHILE)
	p.expect(token.LPAREN)
	cond := p.ParseExpr()
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.WhileStmt{Token: tok, Cond: cond, Body: body}
}

func (p *parserImpl) parseBlock() *ast.BlockStmt {
	tok := p.token
	p.expect(token.LBRACE)
	stmts := []ast.Statement{}
	for p.token.Type != token.RBRACE && p.token.Type != token.EOF {
		stmt := p.safeParseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	p.expect(token.RBRACE)
	return &ast.BlockStmt{Token: tok, Statements: stmts}
}

func (p *parserImpl) parseReturn() *ast.ReturnStmt {
	// return;
	// return xxx;
	tok := p.token
	p.expect(token.RETURN)
	var expr ast.Expression
	if p.token.Type != token.SEMI {
		expr = p.ParseExpr()
	}
	p.expect(token.SEMI)
	return &ast.ReturnStmt{Token: tok, Value: expr}
}

func (p *parserImpl) expect(kind token.TokenType) {
	if p.token.Type == kind {
		p.nextToken()
		return
	}

	pos := p.token.Offset
	if p.token.Type == token.EOF {
		pos = p.prevToken.Offset + len(p.prevToken.Lit)
	}
	line, col := p.l.LineMap(pos)

	p.r.Report(diagnostics.ErrExpectedToken, line, col, string(kind), p.token.Lit)
	panic(&parsingError{"unknown token: " + p.token.Lit})
}

type parsingError struct {
	msg string
}
