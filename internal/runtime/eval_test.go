package runtime

import (
	"mk/internal/diagnostics"
	"mk/internal/lexer"
	"mk/internal/parser"
	"testing"
)

func TestEvaluator_Eval(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "add",
			line: `-1>2?"a":"b";`,
			want: `b`,
		},
		{
			name: "eval add",
			line: `let add = fn(x,y){return x+y;};
			add(1,2);
			`,
			want: `3`,
		},
		{
			name: "eval add",
			line: `let add = fn(x,y){ c = x+y; return c*c;};
			add(1,2);
			`,
			want: `9`,
		},
		{
			name: "eval let plus",
			line: `let a = 10+20;`,
			want: `30`,
		},
		{
			name: "eval let plus",
			line: `let b = 20;let a = 10+b;`,
			want: `30`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := parser.NewParser(l, r)
			program := p.ParseCode()
			env := NewEnv(nil)
			e := NewEvaluator(env, r)
			res := e.Eval(program)
			if len(r.Diagnostics()) > 0 {
				for _, d := range r.Diagnostics() {
					t.Fatal(d.String())
				}
			}
			if res.Inspect() != tt.want {
				t.Errorf("Format() = \n%v\n, want \n%v", res.Inspect(), tt.want)
			}
		})
	}
}
