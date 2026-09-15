package opcode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPop

	OpCall
)

type Definition struct {
	Name  string
	Width []int
}

func (i Instructions) String() string {
	var sb strings.Builder

	return sb.String()
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
	OpSub:      {"OpSub", []int{}},
	OpMul:      {"OpMul", []int{}},
	OpDiv:      {"OpDiv", []int{}},
	OpMod:      {"OpMod", []int{}},
	OpPop:      {"OpPop", []int{}},

	OpCall: {"OpCall", []int{1}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("Opcode %d undefined", op)
	}
	return def, nil
}

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
		}
		offset += w
	}
	return inst
}

func ReadUint16(inst Instructions) uint16 {
	return binary.BigEndian.Uint16(inst)
}
