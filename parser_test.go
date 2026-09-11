package mk

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParser_parseLet(t *testing.T) {
	input := `
	let x = 5;
	let y = 10;
	let foo = "bar";
	let a = fn (x, y) { y = y * y;x = x + y;return x+y;};
	`
	l := NewLexer(input)
	p := NewParser(l)
	program := p.parseProgram()
	if program == nil {
		t.Fatalf("parseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("parseProgram() got= %d, want = %d", len(program.Statements), 3)
	}
	tests := []struct {
		name string
		want *LetStmt
	}{
		{
			name: "",
			want: &LetStmt{},
		},
	}
	for i, _ := range tests {
		stmt := program.Statements[i]
		if stmt.Text() != "let" {
			t.Errorf("")
		}
		let, ok := stmt.(*LetStmt)
		if !ok {
			t.Errorf("statement is not LetStmt. got = %T", let)
		}

	}
}

func TestParser_parseIf(t *testing.T) {
	input := `
	if (x < y) {
	  a = 10;
	} else {
      a = 20;
	}
	`
	l := NewLexer(input)
	p := NewParser(l)
	program := p.parseProgram()
	if program == nil {
		t.Fatalf("parseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("parseProgram() got= %d, want = %d", len(program.Statements), 3)
	}
	tests := []struct {
		want string
	}{
		{"x"},
		{"y"},
		{"foo"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if stmt.Text() != "let" {
			t.Errorf("")
		}
		let, ok := stmt.(*IfStmt)
		if !ok {
			t.Errorf("statement is not LetStmt. want = %s, got = %T", tt.want, let)
		}
	}
}

func Test_parserImpl_parseWhile(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *WhileStmt
		err  error
	}{
		{
			name: "cond is nil",
			line: `while(){}`,
			want: &WhileStmt{Cond: nil, Body: &BlockStmt{}},
			err:  fmt.Errorf("Syntax error"),
		},
		{
			name: "cond != nil",
			line: `while(a != 10){}`,
			want: &WhileStmt{Cond: &BinaryExpr{Lhs: &IdentExpr{Value: "a"}, Op: Token{Lit: "!=", Type: NE}, Rhs: &LiteralExpr{Value: "10"}}, Body: &BlockStmt{}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(NewLexer(tt.line))
			got, gotErr := p.parseWhile()
			if gotErr != nil {
				if tt.err == nil {
					t.Errorf("parseFor() failed: %v, %T, name: %s", gotErr, got, tt.name)
				} else {
					return
				}
			}
			if reflect.TypeOf(got.Cond) != reflect.TypeOf(tt.want.Cond) {
				t.Errorf("parseFor() failed, name = %s, got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_parserImpl_parseReturn(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *ReturnStmt
		err  error
	}{
		{
			name: "int",
			line: `return 10;`,
			want: &ReturnStmt{Value: &LiteralExpr{Value: "10"}},
			err:  nil,
		},
		{
			name: "string",
			line: `return "tom";`,
			want: &ReturnStmt{Value: &LiteralExpr{Value: "tom"}},
			err:  nil,
		},
		{
			name: "void",
			line: `return;`,
			want: &ReturnStmt{Value: nil},
			err:  nil,
		},
		{
			name: "function",
			line: `return fn(x,y) {};`,
			want: &ReturnStmt{Value: &FnExpr{Args: []*IdentExpr{}, Body: &BlockStmt{}}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(NewLexer(tt.line))
			got, gotErr := p.parseReturn()
			if gotErr != nil {
				t.Errorf("parseReturn() failed: %v", gotErr)
			}
			if got.Value != tt.want.Value {
				t.Errorf("parseReturn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parserImpl_parseBlock(t *testing.T) {}

func Test_parserImpl_parseClass(t *testing.T) {}

func Test_parserImpl_parseLet(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *LetStmt
		err  error
	}{
		{
			name: "int",
			line: `let a = 10`,
			want: &LetStmt{Name: &IdentExpr{Value: "a"}, Value: &LiteralExpr{Value: "10"}},
			err:  nil,
		},
		{
			name: "string",
			line: `let a = "tom"`,
			want: &LetStmt{Name: &IdentExpr{Value: "a"}, Value: &LiteralExpr{Value: "tom"}},
			err:  nil,
		},
		{
			name: "function",
			line: `let a = fn(x,y) {}`,
			want: &LetStmt{Name: &IdentExpr{Value: "a"}, Value: &LiteralExpr{Value: "tom"}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(NewLexer(tt.line))
			got, gotErr := p.parseLet()
			if gotErr != nil {
				t.Errorf("parseLet() failed: %v", gotErr)
			}
			if reflect.TypeOf(got.Name) != reflect.TypeOf(tt.want.Name) && reflect.TypeOf(got.Value) != reflect.TypeOf(tt.want.Value) {
				t.Errorf("parseLet() failed, name = %s, got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}
