package interp

import (
	"testing"

	"mk/internal/diagnostics"
	"mk/internal/parser"
	"mk/internal/sym"
	"mk/pkg/util"
)

func newTestEvaluator(t *testing.T) *Evaluator {
	t.Helper()
	ctx := util.NewContext()
	global := sym.NewRootTable(ctx)
	r := diagnostics.NewReporter("", "")
	return NewEvaluator(global, r)
}

func runOn(t *testing.T, e *Evaluator, code string) string {
	t.Helper()

	r := diagnostics.NewReporter("", code)
	l := parser.NewLexer(r)
	l.Reset("", code)
	p := parser.NewParser(l, r)
	program := p.ParseCode()

	if diags := r.Diagnostics(); len(diags) > 0 {
		for _, d := range diags {
			t.Fatalf("parse error: %s", d.Message)
		}
	}

	res := e.Eval(program)

	if diags := r.Diagnostics(); len(diags) > 0 {
		for _, d := range diags {
			t.Fatalf("eval error: %s", d.Message)
		}
	}
	return inspect(res)
}

func runEval(t *testing.T, code string) string {
	t.Helper()
	return runOn(t, newTestEvaluator(t), code)
}

func TestEvaluator_Values(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"arithmetic precedence", "1 + 2 * 3;", "7"},
		{"unary minus", "-5 + 2;", "-3"},
		{"unary bang", "!false;", "true"},
		{"bitwise not", "~0;", "-1"},
		{"string concat", `"foo" + "bar";`, "foobar"},
		{"string compare", `"a" < "b";`, "true"},
		{"int equality", "1 == 1;", "true"},
		{"string equality", `"a" == "a";`, "true"},
		{"let and ident", "let a = 10 + 20; a;", "30"},
		{"assign", "let a = 1; a = 5; a;", "5"},
		{"ternary", `1 > 2 ? "a" : "b";`, "b"},
		{"if else", "let r = 0; if (1 > 0) { r = 1; } else { r = 2; } r;", "1"},
		{"while", "let i = 0; let s = 0; while (i < 5) { s = s + i; i = i + 1; } s;", "10"},
		{"list index", "let l = [1, 2, 3]; l[1];", "2"},
		{"list assign", "let l = [1, 2, 3]; l[1] = 9; l[1];", "9"},
		{"map index", `let m = {"a": 1, "b": 2}; m["b"];`, "2"},
		{"map assign", `let m = {"a": 1}; m["a"] = 5; m["a"];`, "5"},
		{"map dot", `let m = {"a": 1}; m.a;`, "1"},
		{"map dot assign", `let m = {"a": 1}; m.a = 5; m.a;`, "5"},
		{"builtin len", "len([1, 2, 3]);", "3"},
		{"builtin len string", `len("hello");`, "5"},
		{"builtin sizeof", "sizeof(42);", "8"},
		{"user case: let+fn+call", "let a= 10; let b = 20; let c = fn(x, y) {return x+y;}; let d = c(a,b); d;", "30"},
		{"assert pass", "let a = 1; assert(a == 1); a;", "1"},
		{"let redefine", "let a = 1; let a = 2; a;", "2"},
		{"builtin typeof", "typeof(42);", "INT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runEval(t, tt.code); got != tt.want {
				t.Errorf("Eval(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestEvaluator_Functions(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"call", "let add = fn(x, y) { return x + y; }; add(1, 2);", "3"},
		{"implicit return", "let add = fn(x, y) { x + y; }; add(2, 3);", "5"},
		{"closure", "let adder = fn(x) { return fn(y) { return x + y; }; }; let add5 = adder(5); add5(3);", "8"},
		{"recursion", "let fib = fn(n) { if (n < 2) { return n; } return fib(n - 1) + fib(n - 2); }; fib(10);", "55"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runEval(t, tt.code); got != tt.want {
				t.Errorf("Eval(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

// TestEvaluator_ReplPersistence 验证同一个 Evaluator 跨多次 Eval 时 env/symtab 仍然有效。
func TestEvaluator_ReplPersistence(t *testing.T) {
	e := newTestEvaluator(t)

	if got := runOn(t, e, "let g = 5;"); got != "5" {
		t.Fatalf("line1 = %v, want 5", got)
	}
	if got := runOn(t, e, "g * 100;"); got != "500" {
		t.Fatalf("line2 = %v, want 500", got)
	}
	if got := runOn(t, e, "let square = fn(x) { return x * x; }; square(g);"); got != "25" {
		t.Fatalf("line3 = %v, want 25", got)
	}
}

// TestEvaluator_AssertFailed 验证 assert 条件为假时会上报 E4006 运行时诊断，且整体结果为 nil。
func TestEvaluator_AssertFailed(t *testing.T) {
	code := "let a = 1; assert(a == 2);"

	r := diagnostics.NewReporter("", code)
	ctx := util.NewContext()
	e := NewEvaluator(sym.NewRootTable(ctx), r)

	l := parser.NewLexer(r)
	l.Reset("", code)
	p := parser.NewParser(l, r)
	prog := p.ParseCode()
	if ds := r.Diagnostics(); len(ds) > 0 {
		t.Fatalf("parse error: %v", ds[0].Message)
	}

	if res := e.Eval(prog); res != nil {
		t.Fatalf("Eval = %v, want nil", res)
	}

	for _, d := range r.Diagnostics() {
		if d.Descriptor.Code == "E4006" {
			return
		}
	}
	t.Fatalf("expected E4006 diagnostic, got %v", r.Diagnostics())
}
