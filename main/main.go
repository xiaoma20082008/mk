package main

import (
	"flag"
	"fmt"
	"mk/internal/repl"
	"os"
	"os/user"
)

var (
	lspPtr  = flag.Bool("lsp", false, "Run in LSP mode")
	replPtr = flag.Bool("repl", false, "Run in REPL mode")
	helpPtr = flag.Bool("help", false, "Print the usage")
)

func main() {
	flag.Parse()
	if *lspPtr {
		runLsp()
	} else {
		runRepl()
	}
}

func runRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf(`Hello %s, This is the Monkey programming language !

Feel Free to type in commands.
`, user.Username)
	repl.Start(os.Stdin, os.Stdout)
}

func runLsp() {
	// TODO
	fmt.Println(`The Monkey's language server is currently under development, Thank you!`)
}
