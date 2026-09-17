package repl

import (
	"fmt"
	"io"
	"mk/internal/analyzer"
	"mk/internal/diagnostics"
	"mk/internal/interp"
	"mk/internal/lexer"
	"mk/internal/parser"
	"mk/internal/runtime"
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

	// 两个引擎都是有状态的（树遍历持有 env，VM 持有符号表与全局区），
	// 因此长期持有实例、跨行复用；/vm 只在两者之间切换，不重建。
	env := runtime.NewEnv(nil)
	engines := map[interp.Kind]interp.Engine{
		interp.KindTree: interp.New(interp.KindTree, env),
		interp.KindVM:   interp.New(interp.KindVM, env),
	}
	kind := interp.KindVM
	engine := engines[kind]

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
		if line == "/vm" {
			if kind == interp.KindVM {
				kind = interp.KindTree
			} else {
				kind = interp.KindVM
			}
			engine = engines[kind]
			fmt.Fprintf(terminal, "Switched to %s.\n", engine.Name())
			continue
		}

		// 5. 开始执行：lex -> parse ->（compile）-> engine
		execute(engine, "repl", line, terminal)
	}
}

// execute 走完整链路：词法 -> 语法 -> 语义分析（占位）-> 交给引擎执行。
// 引擎内部是否再编译成字节码，由具体实现决定，这一层不关心。
func execute(engine interp.Engine, file, code string, out io.Writer) {
	// 1. 词法与语法分析
	r := diagnostics.NewReporter(file, code)
	l := lexer.NewLexer(code, r)
	p := parser.NewParser(l, r)
	program := p.ParseCode()
	if len(r.Diagnostics()) > 0 {
		printDiagnostics(r, out)
		return
	}

	// 2. 语义分析 TODO（analyzer.Resolve 目前是空实现，名字解析实际发生在编译期）
	analyzer.New().Resolve(program)

	// 3. 交给执行引擎
	res, err := engine.Exec(program, r)
	if err != nil {
		fmt.Fprintln(out, "[Runtime Error]", err)
		return
	}
	if len(r.Diagnostics()) > 0 {
		printDiagnostics(r, out)
		return
	}

	// 4. 输出结果
	if res != nil {
		fmt.Fprintln(out, res.Inspect())
	}
}

func printDiagnostics(r diagnostics.DiagnosticReporter, out io.Writer) {
	for _, d := range r.Diagnostics() {
		fmt.Fprintln(out, d.String())
	}
}
