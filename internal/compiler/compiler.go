package compiler

import (
	"fmt"

	"mk/internal/ast"
	"mk/internal/oop"
	"mk/internal/opcode"
	"mk/internal/sym"
	"mk/internal/token"
)

type Compiler interface {
	Compile(node ast.Node) (constants []oop.Obj, bytecodes opcode.Instructions, err error)
}

// emittedInstruction 记录最近两条指令，用于回填跳转偏移、以及把函数体末尾的 OpPop 改写成 OpReturnValue
type emittedInstruction struct {
	Opcode   opcode.Opcode
	Position int
}

// compilationScope 一段独立的指令流。主程序是第 0 层，每个函数字面量会新开一层。
type compilationScope struct {
	instructions        opcode.Instructions
	lastInstruction     emittedInstruction
	previousInstruction emittedInstruction
}

type compiler struct {
	scopes      []compilationScope
	scopeIndex  int
	constants   []oop.Obj
	symbolTable *sym.SymbolTable
}

func NewCompiler() Compiler {
	return &compiler{
		scopes:      []compilationScope{{instructions: opcode.Instructions{}}},
		scopeIndex:  0,
		constants:   []oop.Obj{},
		symbolTable: sym.NewSymbolTable(),
	}
}

// NewCompilerWithState 复用已有的符号表与常量池（REPL 场景下跨行保持全局状态）
func NewCompilerWithState(s *sym.SymbolTable, constants []oop.Obj) Compiler {
	c := NewCompiler().(*compiler)
	if s != nil {
		c.symbolTable = s
	}
	if constants != nil {
		c.constants = constants
	}
	return c
}

func (c *compiler) SymbolTable() *sym.SymbolTable {
	return c.symbolTable
}

func (c *compiler) Compile(node ast.Node) ([]oop.Obj, opcode.Instructions, error) {
	if node == nil {
		return nil, nil, nil
	}
	_, err := node.Accept(c, nil)
	if err != nil {
		return nil, nil, nil
	}
	return c.constants, c.currentInstructions(), nil
}

// ------------------------------------------------------------------------------------------
// 指令发射
// ------------------------------------------------------------------------------------------

func (c *compiler) emit(op opcode.Opcode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *compiler) addInstruction(ins opcode.Instructions) int {
	scope := &c.scopes[c.scopeIndex]
	pos := len(scope.instructions)
	scope.instructions = append(scope.instructions, ins...)
	return pos
}

func (c *compiler) setLastInstruction(op opcode.Opcode, pos int) {
	scope := &c.scopes[c.scopeIndex]
	scope.previousInstruction = scope.lastInstruction
	scope.lastInstruction = emittedInstruction{Opcode: op, Position: pos}
}

func (c *compiler) lastInstructionIs(op opcode.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *compiler) removeLastPop() {
	scope := &c.scopes[c.scopeIndex]
	scope.instructions = scope.instructions[:scope.lastInstruction.Position]
	scope.lastInstruction = scope.previousInstruction
}

func (c *compiler) replaceInstruction(pos int, newIns opcode.Instructions) {
	scope := &c.scopes[c.scopeIndex]
	for i := 0; i < len(newIns); i++ {
		scope.instructions[pos+i] = newIns[i]
	}
}

// changeOperand 回填跳转指令的目标地址
func (c *compiler) changeOperand(opPos int, operand int) {
	op := opcode.Opcode(c.scopes[c.scopeIndex].instructions[opPos])
	c.replaceInstruction(opPos, opcode.Make(op, operand))
}

func (c *compiler) replaceLastPopWithReturn() {
	pos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(pos, opcode.Make(opcode.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = opcode.OpReturnValue
}

func (c *compiler) addConstant(obj oop.Obj) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *compiler) currentInstructions() opcode.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *compiler) enterScope() {
	c.scopes = append(c.scopes, compilationScope{instructions: opcode.Instructions{}})
	c.scopeIndex++
	c.symbolTable = sym.NewEnclosedSymbolTable(c.symbolTable)
}

func (c *compiler) leaveScope() opcode.Instructions {
	ins := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer()
	return ins
}

func (c *compiler) loadSymbol(s sym.Symbol) {
	switch s.Scope {
	case sym.GlobalScope:
		c.emit(opcode.OpGetGlobal, s.Index)
	case sym.LocalScope:
		c.emit(opcode.OpGetLocal, s.Index)
	case sym.BuiltinScope:
		c.emit(opcode.OpGetBuiltin, s.Index)
	case sym.FreeScope:
		c.emit(opcode.OpGetFree, s.Index)
	}
}

func (c *compiler) storeSymbol(s sym.Symbol) {
	switch s.Scope {
	case sym.GlobalScope:
		c.emit(opcode.OpSetGlobal, s.Index)
	case sym.LocalScope:
		c.emit(opcode.OpSetLocal, s.Index)
	case sym.FreeScope:
		c.emit(opcode.OpSetFree, s.Index)
	case sym.BuiltinScope:
		// 内建函数不可被重新赋值，这里退化成「仅弹出赋值结果」
		c.emit(opcode.OpPop)
	}
}

func (c *compiler) compile(node ast.Node) error {
	if node == nil {
		return nil
	}
	_, err := node.Accept(c, nil)
	return err
}

// ------------------------------------------------------------------------------------------
// Program & Stmt
// ------------------------------------------------------------------------------------------

func (c *compiler) VisitProgram(n *ast.Program, x any) (ast.Node, error) {
	// 预声明顶层 let，使全局名字可以被前向引用（与树遍历解释器的行为保持一致）
	for _, stmt := range n.Statements {
		if let, ok := stmt.(*ast.LetStmt); ok {
			if _, ok := c.symbolTable.Resolve(let.Name.Value); !ok {
				c.symbolTable.Define(let.Name.Value)
			}
		}
	}
	for _, stmt := range n.Statements {
		if err := c.compile(stmt); err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (c *compiler) VisitBlock(n *ast.BlockStmt, x any) (ast.Node, error) {
	for _, stmt := range n.Statements {
		if err := c.compile(stmt); err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (c *compiler) VisitLet(n *ast.LetStmt, x any) (ast.Node, error) {
	if err := c.compile(n.Value); err != nil {
		return nil, err
	}
	c.storeSymbol(c.symbolTable.Define(n.Name.Value))
	return n, nil
}

func (c *compiler) VisitExpr(n *ast.ExprStmt, x any) (ast.Node, error) {
	if err := c.compile(n.Expr); err != nil {
		return nil, err
	}
	c.emit(opcode.OpPop)
	return n, nil
}

func (c *compiler) VisitReturn(n *ast.ReturnStmt, x any) (ast.Node, error) {
	if n.Value == nil {
		c.emit(opcode.OpReturn)
		return n, nil
	}
	if err := c.compile(n.Value); err != nil {
		return nil, err
	}
	c.emit(opcode.OpReturnValue)
	return n, nil
}

func (c *compiler) VisitIf(n *ast.IfStmt, x any) (ast.Node, error) {
	if err := c.compile(n.Cond); err != nil {
		return nil, err
	}
	// 条件为假时跳到 else（没有 else 就跳到整个语句之后）
	jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 0)

	if err := c.compile(n.Then); err != nil {
		return nil, err
	}

	if n.Else != nil {
		// then 分支执行完要跳过 else 分支
		jumpPos := c.emit(opcode.OpJump, 0)
		c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))
		if err := c.compile(n.Else); err != nil {
			return nil, err
		}
		c.changeOperand(jumpPos, len(c.currentInstructions()))
	} else {
		c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))
	}
	return n, nil
}

func (c *compiler) VisitWhile(n *ast.WhileStmt, x any) (ast.Node, error) {
	loopStart := len(c.currentInstructions())

	if err := c.compile(n.Cond); err != nil {
		return nil, err
	}
	jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 0)

	if err := c.compile(n.Body); err != nil {
		return nil, err
	}
	c.emit(opcode.OpJump, loopStart)
	c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))
	return n, nil
}

// ------------------------------------------------------------------------------------------
// Expr
// ------------------------------------------------------------------------------------------

func (c *compiler) VisitInt(n *ast.IntLitExpr, x any) (ast.Node, error) {
	c.emit(opcode.OpConstant, c.addConstant(oop.NewInt(n.Value)))
	return n, nil
}

func (c *compiler) VisitString(n *ast.StringLitExpr, x any) (ast.Node, error) {
	c.emit(opcode.OpConstant, c.addConstant(oop.NewString(n.Value)))
	return n, nil
}

func (c *compiler) VisitBool(n *ast.BoolLitExpr, x any) (ast.Node, error) {
	if n.Value {
		c.emit(opcode.OpTrue)
	} else {
		c.emit(opcode.OpFalse)
	}
	return n, nil
}

func (c *compiler) VisitIdent(n *ast.IdentExpr, x any) (ast.Node, error) {
	sym, ok := c.symbolTable.Resolve(n.Value)
	if !ok {
		return nil, fmt.Errorf("undefined variable: %s", n.Value)
	}
	c.loadSymbol(sym)
	return n, nil
}

func (c *compiler) VisitParen(n *ast.ParenExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Expr); err != nil {
		return nil, err
	}
	return n, nil
}

func (c *compiler) VisitUnary(n *ast.UnaryExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Right); err != nil {
		return nil, err
	}
	switch n.Op.Type {
	case token.PLUS:
		// +x 就是 x 本身
	case token.MINUS:
		c.emit(opcode.OpMinus)
	case token.BANG:
		c.emit(opcode.OpBang)
	case token.TILDE:
		c.emit(opcode.OpTilde)
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", n.Op.Lit)
	}
	return n, nil
}

func (c *compiler) VisitBinary(n *ast.BinaryExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Lhs); err != nil {
		return nil, err
	}
	if err := c.compile(n.Rhs); err != nil {
		return nil, err
	}
	switch n.Op.Type {
	case token.PLUS:
		c.emit(opcode.OpAdd)
	case token.MINUS:
		c.emit(opcode.OpSub)
	case token.STAR:
		c.emit(opcode.OpMul)
	case token.SLASH:
		c.emit(opcode.OpDiv)
	case token.PERCENT:
		c.emit(opcode.OpMod)
	case token.LT:
		c.emit(opcode.OpLess)
	case token.LE:
		c.emit(opcode.OpLessEqual)
	case token.GT:
		c.emit(opcode.OpGreater)
	case token.GE:
		c.emit(opcode.OpGreaterEqual)
	case token.EQ:
		c.emit(opcode.OpEqual)
	case token.NE:
		c.emit(opcode.OpNotEqual)
	case token.LTLT:
		c.emit(opcode.OpShiftLeft)
	case token.GTGT:
		c.emit(opcode.OpShiftRight)
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", n.Op.Lit)
	}
	return n, nil
}

func (c *compiler) VisitTernary(n *ast.TernaryExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Cond); err != nil {
		return nil, err
	}
	jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 0)

	if err := c.compile(n.Then); err != nil {
		return nil, err
	}
	jumpPos := c.emit(opcode.OpJump, 0)

	c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))
	if err := c.compile(n.Else); err != nil {
		return nil, err
	}
	c.changeOperand(jumpPos, len(c.currentInstructions()))
	return n, nil
}

func (c *compiler) VisitCall(n *ast.CallExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Fn); err != nil {
		return nil, err
	}
	for _, arg := range n.Args {
		if err := c.compile(arg); err != nil {
			return nil, err
		}
	}
	c.emit(opcode.OpCall, len(n.Args))
	return n, nil
}

func (c *compiler) VisitAssign(n *ast.AssignExpr, x any) (ast.Node, error) {
	switch lhs := n.Lhs.(type) {
	case *ast.IdentExpr:
		if err := c.compile(n.Rhs); err != nil {
			return nil, err
		}
		sym, ok := c.symbolTable.Resolve(lhs.Value)
		if !ok {
			// 变量未声明：就地声明（沿用树遍历解释器的宽松语义）
			sym = c.symbolTable.Define(lhs.Value)
		}
		// 赋值表达式本身也是值：先复制一份，写入变量后再把值留在栈顶
		c.emit(opcode.OpDup)
		c.storeSymbol(sym)
	case *ast.IndexExpr:
		// a[i] = v
		if err := c.compile(lhs.Lhs); err != nil {
			return nil, err
		}
		if err := c.compile(lhs.Index); err != nil {
			return nil, err
		}
		if err := c.compile(n.Rhs); err != nil {
			return nil, err
		}
		// OpSetIndex 会把写入的值留在栈顶，因此不需要 OpDup
		c.emit(opcode.OpSetIndex)
	default:
		return nil, fmt.Errorf("unsupported assignment target: %s", n.Token.Lit)
	}
	return n, nil
}

func (c *compiler) VisitIndex(n *ast.IndexExpr, x any) (ast.Node, error) {
	if err := c.compile(n.Lhs); err != nil {
		return nil, err
	}
	if err := c.compile(n.Index); err != nil {
		return nil, err
	}
	c.emit(opcode.OpIndex)
	return n, nil
}

func (c *compiler) VisitList(n *ast.ListLitExpr, x any) (ast.Node, error) {
	for _, elem := range n.Value {
		if err := c.compile(elem); err != nil {
			return nil, err
		}
	}
	c.emit(opcode.OpList, len(n.Value))
	return n, nil
}

func (c *compiler) VisitTuple(n *ast.TupleLitExpr, x any) (ast.Node, error) {
	for _, elem := range n.Value {
		if err := c.compile(elem); err != nil {
			return nil, err
		}
	}
	c.emit(opcode.OpTuple, len(n.Value))
	return n, nil
}

func (c *compiler) VisitMap(n *ast.MapLitExpr, x any) (ast.Node, error) {
	for k, v := range n.Value {
		if err := c.compile(k); err != nil {
			return nil, err
		}
		if err := c.compile(v); err != nil {
			return nil, err
		}
	}
	c.emit(opcode.OpMap, len(n.Value)*2)
	return n, nil
}

// VisitFn 编译函数字面量：
// 1. 新开作用域与符号表，编译函数体
// 2. 保证函数体一定以返回指令结尾（隐式返回最后一个表达式的值）
// 3. 把捕获到的自由变量逐个压栈，再发射 OpClosure
func (c *compiler) VisitFn(n *ast.FnExpr, x any) (ast.Node, error) {
	c.enterScope()

	for _, arg := range n.Args {
		c.symbolTable.Define(arg.Value)
	}

	if err := c.compile(n.Body); err != nil {
		c.leaveScope()
		return nil, err
	}

	// 函数体末尾是表达式语句时，把它的值作为返回值（隐式 return）
	if c.lastInstructionIs(opcode.OpPop) {
		c.replaceLastPopWithReturn()
	}
	if !c.lastInstructionIs(opcode.OpReturnValue) {
		c.emit(opcode.OpReturn)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.NumDefinitions()
	instructions := c.leaveScope()

	// 回到外层作用域后，自由变量按 FreeSymbols 的顺序压栈，供 OpClosure 捕获
	for _, sym := range freeSymbols {
		c.loadSymbol(sym)
	}

	compiledFn := &sym.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: len(n.Args),
	}
	c.emit(opcode.OpClosure, c.addConstant(compiledFn), len(freeSymbols))
	return n, nil
}

func (c *compiler) VisitDot(n *ast.DotExpr, x any) (ast.Node, error) {
	return nil, fmt.Errorf("member access `.` is not supported by the stack-based vm yet")
}

// Class
func (c *compiler) VisitClass(n *ast.ClassStmt, x any) (ast.Node, error) {
	return nil, fmt.Errorf("class declaration is not supported by the stack-based vm yet")
}
