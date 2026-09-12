package mk

type Compiler interface {
	Compile(node Node) error
	Bytecode() *Bytecode
}

type compiler struct {
	instructions Instructions
	constants    []Obj
}

type Bytecode struct {
	instructions Instructions
	constants    []Obj
}

func NewCompiler() Compiler {
	return &compiler{
		instructions: Instructions{},
		constants:    []Obj{},
	}
}

func (c *compiler) Compile(node Node) error {
	return nil
}

func (c *compiler) Bytecode() *Bytecode {
	return &Bytecode{
		instructions: c.instructions,
		constants:    c.constants,
	}
}
