package mk

import (
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
			line: `-1>2?"a":"b"`,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReporter("", tt.line)
			l := NewLexer(tt.line, r)
			p := NewParser(l, r)
			program := p.ParseCode()
			env := NewEnv(nil)
			e := NewEvaluator(env)
			res := e.Eval(program)
			if res.Inspect() != tt.want {
				t.Errorf("Format() = \n%v\n, want \n%v", res.Inspect(), tt.want)
			}
		})
	}
}
