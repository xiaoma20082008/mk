package mk

import "testing"

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
		let, ok := stmt.(*LetStmt)
		if !ok {
			t.Errorf("statement is not LetStmt. got = %T", let)
		}
		if let.Name.Value != tt.want {
			t.Errorf("LetStmt.Name.Value is not '%s'. got = %s", tt.want, let.Name.Value)
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

func TestParser_parseFor(t *testing.T) {
	input := `

	for(a=20;;){
	}
	`
	l := NewLexer(input)
	p := NewParser(l)
	program := p.parseProgram()
	if program == nil {
		t.Fatalf("parseProgram() returned nil")
	}
	if len(program.Statements) != 4 {
		t.Fatalf("parseProgram() got= %d, want = %d", len(program.Statements), 3)
	}
	tests := []struct {
		want string
	}{
		{`for (x = 0; x < 10; x = x + 1) {
  let a = 10;
  let b = a * x;
  foo(a, b);
}`},
		{`for (;;) {
}`},
		{`for (; a < b; ) {
}`},
		{`for (a = 10; a < 20; a = a + 1) {
}`},
	}
	pretty := NewFormatter()
	for i, tt := range tests {
		stmt := program.Statements[i]
		fs, ok := stmt.(*ForStmt)
		if !ok {
			t.Errorf("statement is not ForStmt. got = %T", fs)
		}
		got := pretty.Format(stmt)
		if got != tt.want {
			t.Errorf("ForStmt error. want = %T got = %T", tt.want, got)
		}
	}
}
