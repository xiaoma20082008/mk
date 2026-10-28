package vm_test

import (
	"io"
	"mk/internal/vm"
	"os"
	"testing"
)

func TestVM_InterpRun(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		code    string
		w       io.Writer
		wantErr bool
	}{
		{
			name: "",
			code: `
let a = 10;
let b = 20;
let c = fn(a,b) {
  return a > b ?: a : b;
};
let d = c(a,b);
`,
			w: os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := vm.New()
			gotErr := vm.InterpRun(tt.code, tt.w)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InterpRun() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InterpRun() succeeded unexpectedly")
			}
		})
	}
}
