package main

import (
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/peterh/liner"
)

var programmerMode = false
var exitRequested = false

var CommaMode commaMode = rationalComma

type commaMode uint8

const (
	undefinedComma = iota
	floatComma
	rationalComma
)

func main() {
	interactive := 1 /* 0: no, 1: maybe, 2: definitely */
	callStack := NewCallStack()

	initfile := os.ExpandEnv("$HOME/.config/cala/rc")
	_, err := os.Stat(initfile)
	if err == nil {
		executeFile(callStack, initfile)
	}

	for _, arg := range os.Args[1:] {
		if arg == "-i" {
			interactive = 2
			continue
		}
		// disable interactive mode if not forced
		if interactive < 2 {
			interactive = 0
		}

		executeFile(callStack, arg)
	}

	if interactive <= 0 {
		return
	}

	ls := liner.NewLiner()
	defer ls.Close()

	varct := 0

	for {
		var prompt [4]byte

		prompt[0] = 'p'

		if programmerMode {
			prompt[0] = 'P'
		}

		switch CommaMode {
		case undefinedComma:
			prompt[1] = '?'
		case floatComma:
			prompt[1] = 'f'
		case rationalComma:
			prompt[1] = 'r'
		}

		prompt[2] = '>'
		prompt[3] = ' '

		line, err := ls.Prompt(string(prompt[:]))
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "%v\n\n", err)
			}
			break
		}

		if line == "help" {
			btnHelp.bval.fn(nil, 0)
			continue
		}

		ls.AppendHistory(line)

		program, perr := parseString(line)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "%v\n\n", perr)
			continue
		}

		vret, eerr := execWithCallStack(program, callStack)
		callStack = callStack[:1]

		if eerr != nil {
			fmt.Fprintf(os.Stderr, "%v\n\n", eerr)
		} else {
			autonumberVar := lookup(callStack, "_autonumber", false, -1)
			prn := func(varname string) bool {
				if autonumberVar.ival.Cmp(&big.Int{}) != 0 {
					fmt.Printf("= %-60s = %s\n\n", vret, varname)
					return true
				} else {
					fmt.Printf("= %s\n\n", vret)
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
				if vret != nil {
					*autovar = *vret
				} else {
					*autovar = *newZeroVal(IVAL, DECFLV, 0)
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

func executeFile(callStack []CallFrame, arg string) {
	file, err := os.Open(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read %s: %v\n", arg, err)
		return
	}
	defer file.Close()

	program, perr := parse(lex(file))
	if perr != nil {
		fmt.Fprintf(os.Stderr, "Could not parse %s:%v\n", arg, perr)
		return
	}

	vret, eerr := execWithCallStack(program, callStack)
	if eerr != nil {
		fmt.Fprintf(os.Stderr, "Could not execute %s: %v\n", arg, eerr)
		return
	}
	if vret.kind != PVAL {
		fmt.Printf("= %s\n\n", vret)
	}
}
