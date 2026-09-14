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
			io.WriteString(out, "Read error: "+err.Error()+"\n")
			// fmt.Fprintln(terminal, "Read error: "+err.Error())
			continue
		}

		if line == "" {
			continue
		}
		if line == "/exit" {
			fmt.Fprintln(terminal, "Bye bye !")
			break
		}

		r := mk.NewReporter("repl.mk", line)
		l := mk.NewLexer(line, r)
		p := mk.NewParser(l, r)

		program := p.ParseCode()
		if len(r.Diagnostics()) > 0 {
			for _, d := range r.Diagnostics() {
				fmt.Fprintln(terminal, d.String())
			}
			continue
		}
		// fmt := mk.NewFormatter()
		// io.WriteString(out, fmt.Format(program))

		res := mk.NewEvaluator(env, r).Eval(program)

		if res != nil {
			fmt.Fprintln(terminal, res.Inspect())
		}
	}

}
