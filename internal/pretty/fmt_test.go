package pretty

import (
	"mk/internal/diagnostics"
	"mk/internal/lexer"
	"mk/internal/parser"
	"testing"
)

func TestPrettyFormatter_Format(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "fn",
			line: `let a = fn (x, y) { y = y * y;x = x + y;return x+y;};`,
			want: `let a = fn(x, y) {
  y = y * y;
  x = x + y;
  return x + y;
};`,
		},

		{
			name: "while",
			line: `while(a<20){}`,
			want: `while (a < 20) {
}`,
		},
		{
			name: "while1",
			line: `while(a<20){ a= a+1; add(a,10);}`,
			want: `while (a < 20) {
  a = a + 1;
  add(a, 10);
}`,
		},
		{
			name: "paren",
			line: `(10);`,
			want: `(10);`,
		},
		{
			name: "let tuple",
			line: `let a=(10,"20",false,);`,
			want: `let a = (10, "20", false,);`,
		},
		{
			name: "let list",
			line: `let a= [10,"20",false];`,
			want: `let a = [10, "20", false];`,
		},
		// 	TODO: golang的map是无序的,这里先不判断了
		// {
		// 	name: "let map",
		// 	line: `let a={1:10,"2":false,true:10};`,
		// 	want: `let a = {1: 10, "2": false, true: 10};`,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := diagnostics.NewReporter("", tt.line)
			l := lexer.NewLexer(tt.line, r)
			p := parser.NewParser(l, r)
			program := p.ParseCode()
			f := NewFormatter()
			got := f.Format(program)
			if got != tt.want {
				t.Errorf("Format() = \n[%v], want \n[%v]", got, tt.want)
			}
		})
	}
}
