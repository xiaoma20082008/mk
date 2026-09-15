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

### Phase 8: JIT

- [ ] bytecode
- [ ] JIT interceptor
