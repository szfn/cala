package main

import (
	"fmt"
	"github.com/wendal/readline-go"
	"os"
)

func main() {
	interactive := 1 /* 0: no, 1: maybe, 2: definitely */
	callStack := NewCallStack()
	for _, arg := range os.Args[1:] {
		if arg == "-i" {
			interactive = 2
			continue
		}
		// disable interactive mode if not forced
		if interactive < 2 {
			interactive = 0
		}

		file, err := os.Open(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read %s: %v\n", arg, err)
			continue
		}
		defer file.Close()

		program, perr := parse(lex(file))
		if perr != nil {
			fmt.Fprintf(os.Stderr, "Could not parse %s:%v\n", arg, perr)
			continue
		}

		vret, eerr := execWithCallStack(program, callStack)
		if eerr != nil {
			fmt.Fprintf(os.Stderr, "Could not execute %s: %v\n", arg, eerr)
			continue
		}
		fmt.Printf("= %s\n", vret)

		callStack = callStack[:1]
	}

	if interactive <= 0 {
		return
	}

	prompt := ""

	for {
		line := readline.ReadLine(&prompt)
		if line == nil {
			return
		}

		program, perr := parseString(*line)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", perr)
			continue
		}

		vret, eerr := execWithCallStack(program, callStack)
		if eerr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", eerr)
		} else {
			fmt.Printf("= %s\n", vret)
		}
		callStack = callStack[:1]

	}

	os.Exit(0)
}
