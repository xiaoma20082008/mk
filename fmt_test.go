package mk_test

import (
	"mk"
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
			name: "for1",
			line: `for(a=10;a<20;a = a+1){}`,
			want: `for (a = 10; a < 20; a = a + 1) {
}`,
		},
		{
			name: "for1",
			line: `for(;;){}`,
			want: `for (;;) {
}`,
		},
		{
			name: "for2",
			line: `for(a=10;;){}`,
			want: `for (a = 10; ;) {
}`,
		},
		// 		{
		// 			name: "for2",
		// 			line: `for(a=10;;){}`,
		// 			want: `for (a = 10; ;) {
		// }`,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := mk.NewLexer(tt.line)
			p := mk.NewParser(l)
			program := p.ParseCode()
			f := mk.NewFormatter()
			got := f.Format(program)
			if got != tt.want {
				t.Errorf("Format() = \n%v\n, want \n%v", got, tt.want)
			}
		})
	}
}
