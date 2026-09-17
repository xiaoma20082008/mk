package opcode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func Make(op Opcode, operands ...int) Instructions {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return []byte{}
	}

	size := 1
	for _, w := range def.Width {
		size += w
	}

	inst := make([]byte, size)
	inst[0] = byte(op)
	offset := 1
	for i, o := range operands {
		w := def.Width[i]
		switch w {
		case 2:
			binary.BigEndian.PutUint16(inst[offset:], uint16(o))
		case 1:
			inst[offset] = byte(o)
		}
		offset += w
	}
	return inst
}

// ReadOperands 从指令流 ins 中按 def 声明的宽度读取操作数，
// 返回操作数列表与读取的字节数（不含操作码本身）。
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.Width))
	offset := 0
	for i, width := range def.Width {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint8(ins Instructions) uint8 {
	return ins[0]
}

func ReadUint16(inst Instructions) uint16 {
	return binary.BigEndian.Uint16(inst)
}

// String 把指令流反汇编成可读文本：每行「偏移 操作码 操作数...」
func (i Instructions) String() string {
	var sb strings.Builder

	for offset := 0; offset < len(i); {
		op := Opcode(i[offset])
		def, err := Lookup(byte(op))
		if err != nil {
			sb.WriteString(fmt.Sprintf("ERROR: %s\n", err))
			continue
		}
		size := 1
		for _, w := range def.Width {
			size += w
		}
		if offset+size > len(i) {
			sb.WriteString(fmt.Sprintf("ERROR: truncated instruction %s at %d\n", def.Name, offset))
			break
		}
		operands, read := ReadOperands(def, i[offset+1:])
		sb.WriteString(fmt.Sprintf("%04d %s", offset, def.Name))
		for _, operand := range operands {
			sb.WriteString(fmt.Sprintf(" %d", operand))
		}
		sb.WriteString("\n")
		offset += read + 1
	}

	return sb.String()
}
