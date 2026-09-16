package opcode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Instructions []byte

type Opcode byte

const (
	// OpConstant 从常量池加载常量压栈
	OpConstant Opcode = iota
	OpNull
	OpTrue
	OpFalse
	// OpDup 复制栈顶元素，用于「赋值表达式本身也是值」的场景
	OpDup
	OpPop

	// 算术运算
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

	// 一元运算
	OpBang
	OpMinus
	OpTilde

	// 比较运算
	OpEqual
	OpNotEqual
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual

	// 位运算
	OpShiftLeft
	OpShiftRight

	// 控制流
	OpJump
	OpJumpNotTruthy

	// 变量存取
	OpSetGlobal
	OpGetGlobal
	OpSetLocal
	OpGetLocal
	OpSetFree
	OpGetFree
	OpGetBuiltin

	// 调用约定
	OpCall
	OpReturnValue
	OpReturn

	OpNew

	// 闭包
	OpClosure

	// 容器与索引
	OpList
	OpMap
	OpTuple
	OpIndex
	OpSetIndex
)

type Definition struct {
	Name  string
	Width []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpNull:     {"OpNull", []int{}},
	OpTrue:     {"OpTrue", []int{}},
	OpFalse:    {"OpFalse", []int{}},
	OpDup:      {"OpDup", []int{}},
	OpPop:      {"OpPop", []int{}},

	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},
	OpMod: {"OpMod", []int{}},

	OpBang:  {"OpBang", []int{}},
	OpMinus: {"OpMinus", []int{}},
	OpTilde: {"OpTilde", []int{}},

	OpEqual:        {"OpEqual", []int{}},
	OpNotEqual:     {"OpNotEqual", []int{}},
	OpGreater:      {"OpGreater", []int{}},
	OpGreaterEqual: {"OpGreaterEqual", []int{}},
	OpLess:         {"OpLess", []int{}},
	OpLessEqual:    {"OpLessEqual", []int{}},

	OpShiftLeft:  {"OpShiftLeft", []int{}},
	OpShiftRight: {"OpShiftRight", []int{}},

	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	OpSetGlobal:  {"OpSetGlobal", []int{2}},
	OpGetGlobal:  {"OpGetGlobal", []int{2}},
	OpSetLocal:   {"OpSetLocal", []int{1}},
	OpGetLocal:   {"OpGetLocal", []int{1}},
	OpSetFree:    {"OpSetFree", []int{1}},
	OpGetFree:    {"OpGetFree", []int{1}},
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},

	OpCall:        {"OpCall", []int{1}},
	OpReturnValue: {"OpReturnValue", []int{}},
	OpReturn:      {"OpReturn", []int{}},

	OpClosure: {"OpClosure", []int{2, 1}},

	OpList:     {"OpList", []int{2}},
	OpMap:      {"OpMap", []int{2}},
	OpTuple:    {"OpTuple", []int{2}},
	OpIndex:    {"OpIndex", []int{}},
	OpSetIndex: {"OpSetIndex", []int{}},
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

func (o Opcode) String() string {
	def, err := Lookup(byte(o))
	if err != nil {
		return fmt.Sprintf("Opcode(%d)", byte(o))
	}
	return def.Name
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
