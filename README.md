# mk

The monkey programming language written by go

## Run

Running the mk repl by run:

```shell
go run main/main.go
```

## Status

### Phase 1: Lexer

- [x] Lexer

### Phase 2: Parser

- [x] Statement parser: Using Recursive Descent Parser
- [x] Experssion parser: Using [Pratt parser](https://tdop.github.io/)

### Phase 4: Type checker

- [ ] Type Checker based on AST and Visitor Pattern

### Phase 5: Evaluator

- [x] Evaluator: Using AST interceptor

### Phase 5: Object system

- [ ] struct type: int, bool, float, ...
- [ ] reference type: User defined class

### Phase 6: Repl

- [x] `os.Stdout`

### Phase 8: Bytecode VM

- [x] bytecode: 基于栈的指令集（`internal/opcode`）
- [x] compiler: AST -> 字节码（`internal/compiler`）
- [x] stack-based interpreter: 栈式虚拟机（`internal/vm`）
- [ ] closures: 闭包与自由变量捕获（`OpClosure` / `OpGetFree`）
- [x] builtins: `len` / `puts` / `print` / `type` / `first` / `last` / `push`（`internal/builtin`）
- [x] engine: 统一的执行引擎抽象（`internal/interp`）

### Phase 9: JIT

- [ ] JIT interceptor
