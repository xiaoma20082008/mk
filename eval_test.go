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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.line)
			p := NewParser(l)
			program := p.ParseCode()
			env := NewEnv(nil)
			e := NewEvaluator(env)
			r := e.Eval(program)
			if r.Inspect() != tt.want {
				t.Errorf("Format() = \n%v\n, want \n%v", r.Inspect(), tt.want)
			}
		})
	}
}
