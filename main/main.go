package main

import (
	"fmt"
	"mk/internal/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf(`Hello %s, This is the Monkey programming language !

Feel Free to type in commands.
`, user.Username)
	repl.Start(os.Stdin, os.Stdout)
}
