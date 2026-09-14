package repl

import (
	"fmt"
	"io"
	"mk"
	"os"

	"golang.org/x/term"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	// 1. 把标准的输入流（通常是 os.Stdin）切换为终端的 Raw Mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(out, "failed to set terminal to raw mode: %v\n", err)
		return
	}
	// 2. 使用 defer 在 REPL 退出时还原终端状态，否则退出后你的终端会错乱
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 3. 创建一个高级终端抽象层
	terminal := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{in, out}, PROMPT)
	env := mk.NewEnv(nil)
	for {
		// 4. ReadLine 会自动帮你处理 光标移动(← →)、行内插入、退格删除
		line, err := terminal.ReadLine()
		if err != nil {
			if err == io.EOF {
				break // 用户按了 Ctrl+D 优雅退出
			}
			// io.WriteString(out, "Read error: "+err.Error()+"\n")
			fmt.Fprintln(terminal, "Read error: "+err.Error())
			continue
		}

		if line == "" {
			continue
		}
		if line == "/exit" {
			fmt.Fprintln(terminal, "Bye bye !")
			break
		}

		// 5. 开始执行
		astExecute(env, "repl", line, terminal)
	}
}

func astExecute(env *mk.Env, file, code string, out io.Writer) {
	// 1. 词法与语法分析
	r := mk.NewReporter(file, code)
	l := mk.NewLexer(code, r)
	p := mk.NewParser(l, r)
	program := p.ParseCode()
	if len(r.Diagnostics()) > 0 {
		printDiagnostics(r, out)
		return
	}
	// 2. 执行阶段
	evaluator := mk.NewEvaluator(env, r)
	res := evaluator.Eval(program)
	if len(r.Diagnostics()) > 0 {
		printDiagnostics(r, out)
		return
	}
	// 3. 输出结果
	if res != nil {
		fmt.Fprintln(out, res.Inspect())
	}
}

func vmExecute(env *mk.Env, file, code string, out io.Writer) {
	// 1. 词法与语法分析
	r := mk.NewReporter(file, code)
	l := mk.NewLexer(code, r)
	p := mk.NewParser(l, r)
	program := p.ParseCode()
	if len(r.Diagnostics()) > 0 {
		printDiagnostics(r, out)
		return
	}
	// 2. 语义分析 TODO

	// 3. 编译阶段
	c := mk.NewCompiler()
	c.Compile(program)

	// 4. 虚拟机执行阶段
	vm := mk.NewVM(c.Bytecode())
	vm.Run()
}
func printDiagnostics(r mk.DiagnosticReporter, out io.Writer) {
	for _, d := range r.Diagnostics() {
		fmt.Fprintln(out, d.String())
	}
}
