package lexer

import (
	"mk/internal/diagnostics"
	"mk/internal/token"
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
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.STAR, "*"},
		{token.SLASH, "/"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMI, ";"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},

		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMI, ";"},

		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMI, ";"},

		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMI, ";"},
		{token.RBRACE, "}"},

		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMI, ";"},

		{token.STRING, "abc"},
		{token.STRING, " defg 123 "},

		{token.EOF, ""},
	}
	l := NewLexer(input, diagnostics.NewReporter("", input))
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != token.TokenType(tt.Type) {
			t.Fatalf("tests[%d] - tokenType wrong. want=%q, got =%q", i, tt.Type, tok.Type)
		}
		if tok.Lit != tt.Lit {
			t.Fatalf("tests[%d] - tokenLit wrong. want=%q, got =%q", i, tt.Lit, tok.Lit)
		}
	}
}
