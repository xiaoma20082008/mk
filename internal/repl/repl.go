package repl

import (
	"bufio"
	"fmt"
	"io"
	"mk/internal/diagnostics"
	"mk/internal/vm"
	"os"
	"strings"

	"golang.org/x/term"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	// 1.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(out, "failed to set terminal to raw mode: %v\n", err)
		return
	}
	// 2.
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 3.
	terminal := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{in, out}, PROMPT)

	vm := vm.New()
	kind := "ast"

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
			if kind == "ast" {
				kind = "vm"
			} else {
				kind = "ast"
			}
			fmt.Fprintf(terminal, "Switched to %s.\n", kind)
			continue
		}

		// 5. 执行
		switch kind {
		case "ast":
			vm.InterpRun(line, terminal)
		case "vm":
			vm.CompileRun(bufio.NewReaderSize(strings.NewReader(line), 1024), terminal)
		}
	}
}

func printDiagnostics(r diagnostics.DiagnosticReporter, out io.Writer) {
	for _, d := range r.Diagnostics() {
		fmt.Fprintln(out, d.String())
	}
}
