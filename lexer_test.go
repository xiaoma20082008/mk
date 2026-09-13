package mk

import (
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	input := `=+-*/{},;()
	let five = 5;
	let ten = 10;

	let add =  fn(x, y) {
	  return x + y;
	}

	let result = add(five, ten);
	"abc" " defg 123 "
	`
	tests := []struct {
		Type string
		Lit  string
	}{
		{ASSIGN, "="},
		{PLUS, "+"},
		{MINUS, "-"},
		{STAR, "*"},
		{SLASH, "/"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{COMMA, ","},
		{SEMI, ";"},
		{LPAREN, "("},
		{RPAREN, ")"},

		{LET, "let"},
		{IDENT, "five"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMI, ";"},

		{LET, "let"},
		{IDENT, "ten"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMI, ";"},

		{LET, "let"},
		{IDENT, "add"},
		{ASSIGN, "="},
		{FUNCTION, "fn"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COMMA, ","},
		{IDENT, "y"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{SEMI, ";"},
		{RBRACE, "}"},

		{LET, "let"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "five"},
		{COMMA, ","},
		{IDENT, "ten"},
		{RPAREN, ")"},
		{SEMI, ";"},

		{STRING, "abc"},
		{STRING, " defg 123 "},

		{EOF, ""},
	}
	l := NewLexer(input, NewReporter("", input))
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != TokenType(tt.Type) {
			t.Fatalf("tests[%d] - tokenType wrong. want=%q, got =%q", i, tt.Type, tok.Type)
		}
		if tok.Lit != tt.Lit {
			t.Fatalf("tests[%d] - tokenLit wrong. want=%q, got =%q", i, tt.Lit, tok.Lit)
		}
	}
}
