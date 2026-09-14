package mk

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{
			OpConstant, []int{65535}, []byte{byte(OpConstant), 255, 255},
		},
	}

	for _, tt := range tests {
		inst := Make(tt.op, tt.operands...)
		if len(inst) != len(tt.expected) {
			t.Errorf("Instruction has wrong length. want=%d, got=%d", len(tt.expected), len(inst))
		}
		for i, b := range tt.expected {
			if inst[i] != tt.expected[i] {
				t.Errorf("Wrong byte at pos %d. want=%d, got=%d", i, b, inst[i])
			}
		}
	}
}
