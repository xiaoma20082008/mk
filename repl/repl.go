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
		r := mk.NewReporter("", line)
		l := mk.NewLexer(line, r)
		p := mk.NewParser(l, r)

		program := p.ParseCode()
		if len(r.Diagnostics()) > 0 {
			for _, d := range r.Diagnostics() {
				io.WriteString(out, d.String())
				io.WriteString(out, "\n")
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
