package repl

import (
	"bufio"
	"fmt"
	"io"
	"mk"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := mk.NewEnv(nil)
	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			continue
		}
		line := scanner.Text()
		l := mk.NewLexer(line)
		p := mk.NewParser(l)

		program := p.ParseCode()
		if len(p.Errors()) > 0 {
			for _, err := range p.Errors() {
				io.WriteString(out, err.Error())
			}
			continue
		}
		// fmt := mk.NewFormatter()
		// io.WriteString(out, fmt.Format(program))

		res := mk.NewEvaluator(env).Eval(program)

		if res != nil {
			io.WriteString(out, res.Inspect())
			io.WriteString(out, "\n")
		}
		if err := scanner.Err(); err != nil {
			out.Write([]byte(err.Error()))
		}
	}

}
