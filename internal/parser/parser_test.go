package parser

import (
	"fmt"
	"mk/internal/ast"
	"mk/internal/diagnostics"
	"mk/internal/lexer"
	"mk/internal/token"
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
	r := diagnostics.NewReporter("", input)
	l := lexer.NewLexer(input, r)
	p := NewParser(l, r)
	program := p.ParseCode()
	if program == nil {
		t.Fatalf("ParseCode() returned nil")
	}
	if len(r.Diagnostics()) > 0 {
		for _, d := range r.Diagnostics() {
			t.Fatal(d.String())
		}
	}
	if len(program.Statements) != 4 {
		t.Fatalf("ParseCode() got= %d, want = %d", len(program.Statements), 3)
	}
	tests := []struct {
		name string
		want *ast.LetStmt
	}{
		{
			name: "",
			want: &ast.LetStmt{},
		},
	}
	for i, _ := range tests {
		stmt := program.Statements[i]
		if stmt.Text() != "let" {
			t.Errorf("")
		}
		let, ok := stmt.(*ast.LetStmt)
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
	r := diagnostics.NewReporter("", input)
	l := lexer.NewLexer(input, r)
	p := NewParser(l, r)
	program := p.ParseCode()
	if len(r.Diagnostics()) > 0 {
		for _, d := range r.Diagnostics() {
			t.Fatal(d.String())
		}
	}
	tests := []struct {
		want string
	}{
		{"if"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		ifStmt, ok := stmt.(*ast.IfStmt)
		if !ok {
			t.Fatalf("statement is not IfStmt. want = %s, got = %T", tt.want, ifStmt)
		}
	}
}

func Test_parserImpl_parseWhile(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *ast.WhileStmt
		err  error
	}{
		// {
		// 	name: "cond is nil",
		// 	line: `while(){}`,
		// 	want: &WhileStmt{Cond: nil, Body: &BlockStmt{}},
		// 	err:  fmt.Errorf("Syntax error"),
		// },
		// {
		// 	name: "cond != nil",
		// 	line: `while(a != 10){}`,
		// 	want: &WhileStmt{Cond: &BinaryExpr{Lhs: &IdentExpr{Value: "a"}, Op: Token{Lit: "!=", Type: NE}, Rhs: &IntLitExpr{Value: 10}}, Body: &BlockStmt{}},
		// 	err:  nil,
		// },
		{
			name: "cond != nil",
			line: `while(a<20){ a= a+1; add(a,10);}`,
			want: &ast.WhileStmt{Cond: &ast.BinaryExpr{Lhs: &ast.IdentExpr{Value: "a"}, Op: token.Token{Lit: "!=", Type: token.NE}, Rhs: &ast.IntLitExpr{Value: 10}}, Body: &ast.BlockStmt{}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := NewParser(l, r)
			got := p.parseWhile()
			if len(r.Diagnostics()) > 0 {
				for _, d := range r.Diagnostics() {
					t.Error(d.String())
				}
			}
			if reflect.TypeOf(got.Cond) != reflect.TypeOf(tt.want.Cond) {
				t.Errorf("parseWhile() failed, name = %s, got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_parserImpl_parseReturn(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *ast.ReturnStmt
		err  error
	}{
		{
			name: "int",
			line: `return 10;`,
			want: &ast.ReturnStmt{Value: &ast.IntLitExpr{Value: 10}},
			err:  nil,
		},
		{
			name: "string",
			line: `return "tom";`,
			want: &ast.ReturnStmt{Value: &ast.StringLitExpr{Value: "tom"}},
			err:  nil,
		},
		{
			name: "void",
			line: `return;`,
			want: &ast.ReturnStmt{Value: nil},
			err:  nil,
		},
		{
			name: "function",
			line: `return fn(x,y) {};`,
			want: &ast.ReturnStmt{Value: &ast.FnExpr{Args: []*ast.IdentExpr{}, Body: &ast.BlockStmt{}}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := NewParser(l, r)
			got := p.parseReturn()
			if len(r.Diagnostics()) > 0 {
				for _, d := range r.Diagnostics() {
					t.Error(d.String())
				}
			}
			if reflect.TypeOf(got.Value) != reflect.TypeOf(tt.want.Value) {
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
		want *ast.LetStmt
		err  error
	}{
		{
			name: "int",
			line: `let a = 10;`,
			want: &ast.LetStmt{Name: &ast.IdentExpr{Value: "a"}, Value: &ast.IntLitExpr{Value: 10}},
			err:  nil,
		},
		{
			name: "string",
			line: `let a = "tom";`,
			want: &ast.LetStmt{Name: &ast.IdentExpr{Value: "a"}, Value: &ast.StringLitExpr{Value: "tom"}},
			err:  nil,
		},
		{
			name: "bool",
			line: `let a = true;`,
			want: &ast.LetStmt{Name: &ast.IdentExpr{Value: "a"}, Value: &ast.BoolLitExpr{Value: true}},
			err:  nil,
		},
		{
			name: "bool",
			line: `let a = 10 + 20;`,
			want: &ast.LetStmt{Name: &ast.IdentExpr{Value: "a"}, Value: &ast.BinaryExpr{}},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("t.mk", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := NewParser(l, r)
			got := p.parseLet()
			if len(r.Diagnostics()) > 0 {
				for _, d := range r.Diagnostics() {
					fmt.Println(d.String())
				}
				t.Fatal("failed")
			}
			if reflect.TypeOf(got.Name) != reflect.TypeOf(tt.want.Name) && reflect.TypeOf(got.Value) != reflect.TypeOf(tt.want.Value) {
				t.Errorf("parseLet() failed, name = %s, got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_parserImpl_parseExpr(t *testing.T) {
	tests := []struct {
		name string
		line string
		want ast.Expression
	}{
		{
			name: "int",
			line: `10`,
			want: &ast.IntLitExpr{Value: 10},
		},
		{
			name: "string",
			line: `"tom"`,
			want: &ast.StringLitExpr{Value: "tom"},
		},
		{
			name: "bool",
			line: `true`,
			want: &ast.BoolLitExpr{Value: true},
		},
		{
			name: "map",
			line: `{k1:v1}`,
			want: &ast.MapLitExpr{Value: map[ast.Expression]ast.Expression{&ast.IdentExpr{Value: "k1"}: &ast.IdentExpr{Value: "v1"}}},
		},
		{
			name: "map1",
			line: `{k1:v1,}`,
			want: &ast.MapLitExpr{Value: map[ast.Expression]ast.Expression{&ast.IdentExpr{Value: "k1"}: &ast.IdentExpr{Value: "v1"}}},
		},
		{
			name: "map2",
			line: `{k1:v1,k2:v2}`,
			want: &ast.MapLitExpr{Value: map[ast.Expression]ast.Expression{&ast.IdentExpr{Value: "k1"}: &ast.IdentExpr{Value: "v1"}, &ast.IdentExpr{Value: "k2"}: &ast.IdentExpr{Value: "v2"}}},
		},
		{
			name: "map2",
			line: `{k1:v1,k2:v2,}`,
			want: &ast.MapLitExpr{Value: map[ast.Expression]ast.Expression{&ast.IdentExpr{Value: "k1"}: &ast.IdentExpr{Value: "v1"}, &ast.IdentExpr{Value: "k2"}: &ast.IdentExpr{Value: "v2"}}},
		},
		{
			name: "list",
			line: `[v1,v2]`,
			want: &ast.ListLitExpr{Value: []ast.Expression{&ast.IdentExpr{Value: "v1"}, &ast.IdentExpr{Value: "v2"}}},
		},
		{
			name: "tuple1",
			line: `(v1,)`,
			want: &ast.TupleLitExpr{Value: []ast.Expression{&ast.IdentExpr{Value: "v1"}}},
		},
		{
			name: "tuple2",
			line: `(v1,v2,)`,
			want: &ast.TupleLitExpr{Value: []ast.Expression{&ast.IdentExpr{Value: "v1"}, &ast.IdentExpr{Value: "v2"}}},
		},
		{
			name: "group",
			line: `(v1)`,
			want: &ast.ParenExpr{Expr: &ast.IdentExpr{Value: "v1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := NewParser(l, r)
			got := p.ParseExpr()
			if len(r.Diagnostics()) > 0 {
				for _, d := range r.Diagnostics() {
					t.Error(d.String())
				}
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("parseExpr() failed, name = %s, got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_parserImpl_ParseCode(t *testing.T) {
	input := `
	let a = 10;
	let b = "20";
	let add = fn (x, y) { c = a * a; return c *b;};
	`
	r := diagnostics.NewReporter("", input)
	l := lexer.NewLexer(input, r)
	p := NewParser(l, r)
	code := p.ParseCode()

	if len(r.Diagnostics()) > 0 {
		for _, d := range r.Diagnostics() {
			t.Fatal(d.String())
		}
	}
	tests := []struct {
		name string
		want ast.Statement
	}{
		{
			name: "let int",
			want: &ast.LetStmt{},
		},
		{
			name: "let string",
			want: &ast.LetStmt{},
		},
		{
			name: "let fn",
			want: &ast.LetStmt{},
		},
	}
	for i, tt := range tests {
		stmt := code.Statements[i]
		if reflect.TypeOf(stmt) != reflect.TypeOf(tt.want) {
			t.Errorf("ParseCode() failed, name = %s, got = %v, want = %v", tt.name, stmt, tt.want)
		}
	}
}
