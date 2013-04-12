package main

import (
	"bufio"
	"io"
	"fmt"
	"unicode"
)

type tokenType int
const (
	ERRTOK tokenType = iota // an error occoured
	EOFTOK // end of file

	REALTOK  // a real number
	INTTOK // an integer number
	HEXTOK // a hexadecimal number
	OCTTOK // an octal number
	KWDTOK // any builtin keyword
	SYMTOK // any symbol

	PAROPTOK // (
	PARCLTOK // )
	CRLOPTOK // {
	CRLCLTOK // }

	ADDOPTOK // +
	SUBOPTOK // -
	MULOPTOK // *
	DIVOPTOK // /
	MODOPTOK // %
	POWOPTOK // **
	OROPTOK // ||
	BWOROPTOK // |
	ANDOPTOK // &&
	BWANDOPTOK // &
	SETOPTOK // =

	INCOPTOK // ++
	DECOPTOK // --

	NEGOPTOK // !

	ADDEQTOK // +=
	SUBEQTOK // -=
	MULEQTOK // *=
	DIVEQTOK // /=
	MODEQTOK // %=

	EQOPTOK // ==
	GEOPTOK // >=
	GTOPTOK // >
	LEOPTOK // <=
	LTOPTOK // <
	NEOPTOK // !=

	COMMATOK // ,
	SCOLTOK // ;
)

type followCont struct {
	r rune
	ttype tokenType
}

var opTable = map[rune][]followCont{
	'+': []followCont{
		{ '+', INCOPTOK },
		{ '=', ADDEQTOK },
		{ 0, ADDOPTOK }},
	'-': []followCont{
		{ '-', DECOPTOK },
		{ '=', SUBEQTOK },
		{ 0, SUBOPTOK }},
	'*': []followCont{
		{ '*', POWOPTOK },
		{ '=', MULEQTOK },
		{ 0, MULOPTOK }},
	'/': []followCont{
		{ '=', DIVEQTOK },
		{ 0, DIVOPTOK }},
	'%': []followCont{
		{ '=', MODEQTOK },
		{ 0, MODOPTOK }},
	'>': []followCont{
		{ '=', GEOPTOK },
		{ 0, GTOPTOK }},
	'<': []followCont{
		{ '=', LEOPTOK },
		{ 0, LTOPTOK }},
	'!': []followCont{
		{ '=', NEOPTOK },
		{ 0, NEGOPTOK }},
	'|': []followCont{
		{ '|', OROPTOK },
		{ 0, BWOROPTOK }},
	'&': []followCont{
		{ '&', ANDOPTOK },
		{ 0, BWANDOPTOK }},
	'=': []followCont{
		{ '=', EQOPTOK },
		{ 0, SETOPTOK }},
}

var kwdTable = map[string]bool{
	"if": true,
	"else": true,
	"while": true,
	"for": true,
	"func": true,
}

type token struct {
	ttype tokenType
	val string
	lineno int
}

type lexer struct {
	lineno int
	tokStream chan token
	input *bufio.Reader
	acc []rune
	acceptNonsyn bool
}

type lexerStateFn func(lx *lexer) lexerStateFn

func (lx *lexer) emit(ttype tokenType, val string) {
	lx.tokStream <- token{
		ttype,
		val,
		lx.lineno,
	}
}

// If err is not nil emits the appropriate error/eof tokens on tokStream and return true
// otherwise it returns false
// If this function return true you are supposed to return nil and exit
func (lx *lexer) lerror(err error) bool {
	if err != nil {
		if err == io.EOF {
			// suppress eof, the read character will be 0, this is fine
			// the functions that need to handle EOF must do it themselves
		} else {
			lx.emit(ERRTOK, err.Error())
			lx.emit(EOFTOK, "")
			return true
		}
	}
	return false
}


// Returns a syntax error
func (lx *lexer) syntaxError(c rune) {
	lx.emit(ERRTOK, fmt.Sprintf("Syntax error: unexpected character '%c' in line %d", c, lx.lineno))
}



// Reads an operator, remember that operators are at most two characters long. The first character of the operator must have already been read and stored in the accumulator
func lxFollow(lx *lexer) lexerStateFn {
	c, _, err := lx.input.ReadRune()
	if lx.lerror(err) {
		return nil
	}

	continuations, ok := opTable[lx.acc[0]]
	if !ok {
		panic(fmt.Errorf("Internal error: got inside lxFollow with an invalid accumulator: %v", lx.acc))
	}

	for _, cont := range continuations {
		if cont.r == c {
			lx.acc = append(lx.acc, c)
			lx.emit(cont.ttype, string(lx.acc))
			return lxBase
		} else if cont.r == 0 {
			lx.emit(cont.ttype, string(lx.acc))
			return toBase1(lx, c, true)
		}
	}

	lx.syntaxError(c)
	return nil
}

// Reads a comment
func lxComment(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}
		if c == '\n' {
			lx.lineno++
			return lxBase
		}
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads a sequence of alphanumeric characters (a "symbol")
func lxSymbol(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		if unicode.IsLetter(c) || unicode.IsNumber(c) || (c == '_') {
			lx.acc = append(lx.acc, c)
		} else {
			symbol := string(lx.acc)
			ttype := SYMTOK
			if _, ok := kwdTable[symbol]; ok {
				ttype = KWDTOK
			}
			lx.emit(ttype, symbol)
			return toBase1(lx, c, false)
		}
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads a real number
func lxReal(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		if unicode.IsDigit(c) {
			lx.acc = append(lx.acc, c)
		} else if c == '.' {
			lx.acc = append(lx.acc, '.')
			return lxRealFrac
		} else if (c == 'e') || (c == 'E') {
			lx.acc = append(lx.acc, c)
			return lxRealExp
		} else {
			lx.emit(INTTOK, string(lx.acc))
			return toBase1(lx, c, false)
		}
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads the part after the '.' of a real number
func lxRealFrac(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		if unicode.IsDigit(c) {
			lx.acc = append(lx.acc, c)
		} else if (c == 'e') || (c == 'E') {
			lx.acc = append(lx.acc, c)
			return lxRealExp
		} else {
			lx.emit(REALTOK, string(lx.acc))
			return toBase1(lx, c, false)
		}
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads the exponent part of a real number (what comes after the 'e' or 'E')
func lxRealExp(lx *lexer) lexerStateFn {
	first := true

	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		if unicode.IsDigit(c) {
			lx.acc = append(lx.acc, c)
		} else if first && ((c == '+') || (c == '-')) {
			lx.acc = append(lx.acc, c)
		} else {
			lx.emit(REALTOK, string(lx.acc))
			return toBase1(lx, c, false)
		}

		first = false
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads a number, could be an octal number, an hexadecimal number or a fractional number
// We assume that a 0 has already been read and is in lx.acc
func lxNumber(lx *lexer) lexerStateFn {
	c, _, err := lx.input.ReadRune()
	if lx.lerror(err) {
		return nil
	}

	switch c {
	case 'x':
		lx.acc = append(lx.acc, c)
		return lxHex

	case '0', '1', '2', '3', '4', '5', '6', '7':
		lx.acc = append(lx.acc, c)
		return lxOct

	case '.':
		lx.acc = append(lx.acc, c)
		return lxRealFrac

	default: // it was just a zero
		lx.emit(INTTOK, string(lx.acc))
		return toBase1(lx, c, false)
	}

	panic(fmt.Errorf("Unreachable"))
}

// Reads an hexadecimal number
func lxHex(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'a', 'B', 'b', 'C', 'c', 'D', 'd', 'E', 'e', 'F', 'f':
			lx.acc = append(lx.acc, c)
		default:
			lx.emit(HEXTOK, string(lx.acc))
			return toBase1(lx, c, false)
		}
	}
	panic(fmt.Errorf("Unreachable"))
}

// Reads an octal number
func lxOct(lx *lexer) lexerStateFn {
	for {
		c, _, err := lx.input.ReadRune()
		if lx.lerror(err) {
			return nil
		}

		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			lx.acc = append(lx.acc, c)
		default:
			lx.emit(OCTTOK, string(lx.acc))
			return toBase1(lx, c, false)
		}
	}
	panic(fmt.Errorf("Unreachable"))
}



// Helper function, saves c into the accumulator then goes to the specified state
func toState(lx *lexer, c rune, next lexerStateFn) lexerStateFn{
	lx.acc[0] = c
	lx.acc = lx.acc[:1]
	return next
}

// Goes to lxBase1 state after having read character c
func toBase1(lx *lexer, c rune, acceptNonsyn bool) lexerStateFn {
	lx.acceptNonsyn = acceptNonsyn
	return toState(lx, c, lxBase1)
}

// Interprets the one character stored in the accumulator. This is the base state
func lxBase1(lx *lexer) lexerStateFn {
	c := lx.acc[0]

	switch c {
	case 0:
		lx.emit(EOFTOK, "")
		return nil

	case ' ', '\t':
		lx.acceptNonsyn = true
		// skip

	case '\n':
		lx.acceptNonsyn = true
		lx.lineno++

	case '#':
		lx.acceptNonsyn = true
		return lxComment

	case '0':
		if lx.acceptNonsyn  {
			lx.acc = []rune{ c }
			return lxNumber
		} else {
			lx.syntaxError(c)
			return nil
		}

	case '.':
		if lx.acceptNonsyn {
			lx.acc = []rune{ '.' }
			return lxRealFrac
		} else {
			lx.syntaxError(c)
			return nil
		}

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if lx.acceptNonsyn {
			lx.acc = []rune{ c }
			return lxReal
		} else {
			lx.syntaxError(c)
			return nil
		}

	case '(': lx.emit(PAROPTOK, "(")
	case ')': lx.emit(PARCLTOK, ")")
	case '{': lx.emit(CRLOPTOK, "{")
	case '}': lx.emit(CRLCLTOK, "}")
	case ',': lx.emit(COMMATOK, ",")
	case ';': lx.emit(SCOLTOK, ";")

	default:
		if _, ok := opTable[c]; ok {
			return toState(lx, c, lxFollow)
		}

		if (c == '_') || unicode.IsLetter(c) {
			if lx.acceptNonsyn {
				lx.acc = []rune{ c }
				return lxSymbol
			} else {
				lx.syntaxError(c)
				return nil
			}
		}

		lx.syntaxError(c)

		return nil
	}

	return lxBase
}

// Reads one character, stores it into the accumulator and moves to lxBase1 state to interpret it
func lxBase(lx *lexer) lexerStateFn {
	c, _, err := lx.input.ReadRune()

	if lx.lerror(err) {
		return nil
	}

	return toBase1(lx, c, true)
}

// Runs the lexer until the state machine reaches nil
func (lx *lexer) run() {
	for state := lxBase; state != nil; {
		state = state(lx)
	}
	close(lx.tokStream)
}

// Runs the lexer concurrently over input. Returns a tream of tokens
func lex(input io.Reader) (chan token) {
	r := &lexer{
		lineno: 1,
		tokStream: make(chan token),
		input: bufio.NewReader(input),
		acc: make([]rune, 1),
	}
	go r.run()
	return r.tokStream
}

func lexAll(input io.Reader) []token {
	ts := lex(input)
	r := []token{}
	for t := range ts {
		r = append(r, t)
	}
	return r
}

var tokenTypeName = map[tokenType]string{
	ERRTOK: "an error occoured",
	EOFTOK: "end of file",
	REALTOK: "a real number",
	INTTOK: "an integer number",
	HEXTOK: "a hexadecimal number",
	OCTTOK: "an octal number",
	KWDTOK: "any builtin keyword",
	SYMTOK: "any symbol",
	PAROPTOK: "(",
	PARCLTOK: ")",
	CRLOPTOK: "{",
	CRLCLTOK: "}",
	ADDOPTOK: "+",
	SUBOPTOK: "-",
	MULOPTOK: "*",
	DIVOPTOK: "/",
	MODOPTOK: "%",
	POWOPTOK: "**",
	OROPTOK: "||",
	BWOROPTOK: "|",
	ANDOPTOK: "&&",
	BWANDOPTOK: "&",
	SETOPTOK: "=",
	INCOPTOK: "++",
	DECOPTOK: "--",
	NEGOPTOK: "!",
	ADDEQTOK: "+=",
	SUBEQTOK: "-=",
	MULEQTOK: "*=",
	DIVEQTOK: "/=",
	MODEQTOK: "%=",
	EQOPTOK: "==",
	GEOPTOK: "ge",
	GTOPTOK: "gt",
	LEOPTOK: "le",
	LTOPTOK: "lt",
	NEOPTOK: "!=",
	COMMATOK: ",",
	SCOLTOK: ";",
}

func (ttype tokenType) String() string {
	return tokenTypeName[ttype]
}

