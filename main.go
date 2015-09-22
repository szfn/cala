package main

import (
	"fmt"
	"github.com/peterh/liner"
	"io"
	"os"
)

var exitRequested = false

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

	ls := liner.NewLiner()
	defer ls.Close()

	varct := 0

	for {
		line, err := ls.Prompt("")
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			break
		}

		program, perr := parseString(line)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", perr)
			continue
		}

		vret, eerr := execWithCallStack(program, callStack)
		callStack = callStack[:1]

		if eerr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", eerr)
		} else {
			autonumberVar := lookup(callStack, "_autonumber", false, -1)
			prn := func(varname string) bool {
				if autonumberVar.ival != 0 {
					fmt.Printf("= %-60s = %s\n", vret, varname)
					return true
				} else {
					fmt.Printf("= %s\n", vret)
					return false
				}
			}

			if ok, name := isVarLookup(program); ok {
				prn(name)
			} else {
				name = fmt.Sprintf("_%d", varct)
				if prn(name) {
					autovar := lookup(callStack, name, true, -1)
					*autovar = *vret
					varct++
				}
				autovar := lookup(callStack, "_", true, -1)
				fmt.Printf("vret: %v\n", vret)
				if vret != nil {
					*autovar = *vret
				} else {
					*autovar = value{IVAL, 0, 0.0, nil, nil, nil}
				}
			}
		}

		if exitRequested {
			break
		}
	}
}

func isVarLookup(program AstNode) (bool, string) {
	bn, ok := program.(*BodyNode)
	if !ok {
		return false, ""
	}
	if len(bn.statements) != 1 {
		return false, ""
	}
	vn, ok := bn.statements[0].(*VarNode)
	if !ok {
		return false, ""
	}
	return true, vn.name
}
