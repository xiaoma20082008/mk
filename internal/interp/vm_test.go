package interp

import (
	"mk/internal/diagnostics"
	"mk/internal/lexer"
	"mk/internal/parser"
	"testing"
)

func runVM(t *testing.T, e *VMEngine, line string) string {
	t.Helper()
	r := diagnostics.NewReporter("", line)
	l := lexer.NewLexer(line, r)
	p := parser.NewParser(l, r)
	program := p.ParseCode()
	res, err := e.Exec(program, r)
	if err != nil {
		t.Fatalf("Exec(%q) error: %v", line, err)
	}
	if len(r.Diagnostics()) > 0 {
		for _, d := range r.Diagnostics() {
			t.Fatalf("diagnostics: %s", d.String())
		}
	}
	if res == nil {
		return "null"
	}
	return res.Inspect()
}

func TestVMEngine_Basic(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"arithmetic precedence", "1 + 2 * 3;", "7"},
		{"unary minus", "-1 + 5;", "4"},
		{"string concat", `"foo" + "bar";`, "foobar"},
		{"comparison", "3 > 2;", "true"},
		{"let binding", "let a = 10 + 20; a;", "30"},
		{"function call", "let add = fn(x, y) { return x + y; }; add(1, 2);", "3"},
		{"closure free var", "let adder = fn(x) { return fn(y) { return x + y; }; }; let add5 = adder(5); add5(3);", "8"},
		{"builtin len", "len([1, 2, 3]);", "3"},
		{"builtin first", "first([1, 2, 3]);", "1"},
		{"builtin typeof", "typeof(42);", "INT"},
		{"builtin sizeof", "sizeof(42);", "4"},
		{"list literal index", "[1, 2, 3][1];", "2"},
		{"map literal index", `let m = {"a": 1, "b": 2}; m["b"];`, "2"},
		{"list append via push", "let l = [1]; push(l, 2); l[1];", "2"},
		{"while loop", "let i = 0; let s = 0; while (i < 5) { s = s + i; i = i + 1; } s;", "10"},
		{"ternary as assign", `let r = 1 > 2 ? "a" : "b"; r;`, "b"},
		{"if else with assign", "let r = 0; if (2 > 1) { r = 100; } else { r = 200; } r;", "100"},
	}

	e := NewVMEngine()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runVM(t, e, tt.line); got != tt.want {
				t.Errorf("Exec(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestVMEngine_Recursion(t *testing.T) {
	e := NewVMEngine()
	src := `
	let fib = fn(n) {
		if (n < 2) {
			return n;
		}
		return fib(n - 1) + fib(n - 2);
	};
	fib(15);
	`
	if got := runVM(t, e, src); got != "610" {
		t.Errorf("fib(15) = %v, want 610", got)
	}
}

// TestVMEngine_CrossLine 验证 REPL 场景下跨行复用 symbols/constants/globals 的正确性。
func TestVMEngine_CrossLine(t *testing.T) {
	e := NewVMEngine()
	if got := runVM(t, e, "let g = 5;"); got != "5" {
		t.Fatalf("line1 = %v, want 5", got)
	}
	if got := runVM(t, e, "g * 100;"); got != "500" {
		t.Fatalf("line2 = %v, want 500", got)
	}
	if got := runVM(t, e, "let square = fn(x) { return x * x; }; square(g);"); got != "25" {
		t.Fatalf("line3 = %v, want 25", got)
	}
}
