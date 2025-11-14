package main

import (
	"fmt"
	"io"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type tokenStream struct {
	tokStream chan token
	rewound   []token // lookahead tokens
}

func (ts *tokenStream) get() token {
	if len(ts.rewound) > 0 {
		r := ts.rewound[len(ts.rewound)-1]
		ts.rewound = ts.rewound[:len(ts.rewound)-1]
		return r
	}
	r := <-ts.tokStream
	return r
}

func (ts *tokenStream) rewind(t token) {
	ts.rewound = append(ts.rewound, t)
}

type ParseError struct {
	msg        string
	stackTrace []string
}

func getStackTrace() []string {
	trace := []string{}
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		trace = append(trace, fmt.Sprintf("%s:%d", file, line))
	}
	return trace
}

func NewParseError(msg string) *ParseError {
	return &ParseError{msg, getStackTrace()}
}

func (pe *ParseError) Error() string {
	return pe.msg
}

func (pe *ParseError) printStackTrace(out io.Writer) {
	for _, frame := range pe.stackTrace {
		fmt.Fprintf(out, "at %s\n", frame)
	}
}

// Transforms a stream of tokens into a parse tree
func parse(tokStream chan token) (n AstNode, err error) {
	defer func() {
		if p := recover(); p != nil {
			if err2, ok := p.(error); ok {
				err = NewParseError(err2.Error())
			} else {
				panic(p)
			}
		}
	}()
	ts := &tokenStream{tokStream, make([]token, 0)}
	n = parseStatements(ts, true)
	return
}

func parseReader(r io.Reader) (n AstNode, err error) {
	return parse(lex(r))
}

func parseString(s string) (n AstNode, err error) {
	return parse(lex(strings.NewReader(s)))
}

// Parses a list of statements.
// If toplevel is true the function was called directly by parse, toplevel enables:
// - function definitions
// - swallowing eof without throwing a fit
// - enables statements to not be terminated by ';' (normal c-like languages don't have this, we just do it for interactive use)
// statement-list ::= <statement> ; <statement-list>
func parseStatements(ts *tokenStream, toplevel bool) AstNode {
	r := []AstNode{}
	lineno := -1
	for {
		tok := ts.get()
		ts.rewind(tok)

		if lineno < 0 {
			lineno = tok.lineno
		}

		if toplevel {
			if tok.ttype == EOFTOK {
				return NewBodyNode(r, lineno)
			} else if tok.ttype == ERRTOK {
				panic(fmt.Errorf("%s", tok.val))
			}
		} else {
			if tok.ttype == CRLCLTOK {
				return NewBodyNode(r, lineno)
			}
		}

		r = append(r, parseStatement(ts, toplevel))
	}
	panic("Unreachable")
}

// Parses a function definition
func parseFnDef(ts *tokenStream, lineno int) AstNode {
	nameTok := ts.get()
	if nameTok.ttype != SYMTOK {
		unexpectedToken(nameTok, " (while parsing function definition)")
	}

	tokMust(PAROPTOK, ts, " (while parsing function definition)")
	first := true
	args := []string{}
	for {
		tok := ts.get()
		if tok.ttype == PARCLTOK {
			break
		} else if !first {
			if tok.ttype != COMMATOK {
				unexpectedToken(tok, " (expected ',' while parsing function definition)")
			}
			tok = ts.get()
		}

		first = false

		if tok.ttype != SYMTOK {
			unexpectedToken(tok, " (expected symbol while parsing function definition)")
		}
		args = append(args, tok.val)
	}

	tokMust(CRLOPTOK, ts, " (while parsing function definition)")
	body := parseStatements(ts, false)
	tokMust(CRLCLTOK, ts, " (while parsing function definition)")

	return NewFnDefNode(nameTok.val, args, body, lineno)
}

// Parses a statement, the first token already read
// statement ::= <if> | <while> | <for> | <func-def> | "@" [<expression>] | <expression>;
// the semicolon at the end of the expression becomes optional if toplevel == true
// function definitions can only appear at toplevel
func parseStatement(ts *tokenStream, toplevel bool) AstNode {
	tok := ts.get()

	if tok.val == "@" {
		e := parseDpy(ts, tok.lineno)
		parseSemicolon(ts, toplevel)
		return e
	}

	if tok.ttype != KWDTOK {
		ts.rewind(tok)
		e := parseExpressionSet(ts)
		parseSemicolon(ts, toplevel)
		return e
	}

	switch tok.val {
	case "func":
		if !toplevel {
			unexpectedToken(tok, " (can not define nested functions)")
		}
		return parseFnDef(ts, tok.lineno)
	case "if":
		return parseIf(ts, tok.lineno)
	case "while":
		return parseWhile(ts, tok.lineno)
	case "for":
		return parseFor(ts, tok.lineno)
	case "exit":
		e := parseExit(ts, tok.lineno)
		parseSemicolon(ts, toplevel)
		return e
	}
	unexpectedToken(tok, " (while parsing a statement)")
	panic("Unreachable")
}

func parseSemicolon(ts *tokenStream, toplevel bool) {
	tok := ts.get()
	if tok.ttype != SCOLTOK {
		if toplevel {
			ts.rewind(tok)
		} else {
			unexpectedToken(tok, " (expecting ';' while reading statement)")
		}
	}
}

// Parses an if statement, note that the 'if' keyword itself has already been read
// if ::= | if (<expression>) { <statement-list> } | if (<expression>) { <statement-list> } else { <statement-list> } | if (<expression> { <statement-list> } else <if>
func parseIf(ts *tokenStream, lineno int) AstNode {
	tokMust(PAROPTOK, ts, " (parsing 'if' statement)")
	guard := parseExpressionSet(ts)
	tokMust(PARCLTOK, ts, " (parsing 'if' statement)")
	tokMust(CRLOPTOK, ts, " (parsing 'if' statement)")
	ifBody := parseStatements(ts, false)
	tokMust(CRLCLTOK, ts, " (parsing 'if' statement)")
	var elseBody AstNode = nil

	tok := ts.get()
	if (tok.ttype == KWDTOK) && (tok.val == "else") {
		tok2 := ts.get()
		if tok2.ttype == CRLOPTOK {
			elseBody = parseStatements(ts, false)
			tokMust(CRLCLTOK, ts, " (parsing 'else' branch)")
		} else if (tok2.ttype == KWDTOK) && (tok2.val == "if") {
			elseBody = parseIf(ts, tok2.lineno)
		} else {
			unexpectedToken(tok2, " (expected '{' or 'if' parsing 'else' branch)")
		}
	} else {
		ts.rewind(tok)
	}

	return NewIfNode(guard, ifBody, elseBody, lineno)
}

// Parses a while statement, note that the 'while' keyword itself has already been read
// while ::= while (<expression>) { <statement-list> }
func parseWhile(ts *tokenStream, lineno int) AstNode {
	tokMust(PAROPTOK, ts, " (parsing 'while' statement)")
	guard := parseExpressionSet(ts)
	tokMust(PARCLTOK, ts, " (parsing 'while' statement)")
	tokMust(CRLOPTOK, ts, " (parsing 'while' statement)")
	body := parseStatements(ts, false)
	tokMust(CRLCLTOK, ts, " (parsing 'while' statement)")
	return NewWhileNode(guard, body, lineno)
}

// Parses a for statement, note that the 'for' keyword itself has already been read
// for ::= for (<expression>; <expression>; <expression>) { <statement-list> }
func parseFor(ts *tokenStream, lineno int) AstNode {
	tokMust(PAROPTOK, ts, " (parsing 'for' statement)")
	initExpr := parseExpressionSet(ts)
	tokMust(SCOLTOK, ts, " (parsing 'for' statement)")
	guard := parseExpressionSet(ts)
	tokMust(SCOLTOK, ts, " (parsing 'for' statement)")
	incrExpr := parseExpressionSet(ts)
	tokMust(PARCLTOK, ts, " (parsing 'for' statement)")
	tokMust(CRLOPTOK, ts, " (parsing 'for' statement)")
	body := parseStatements(ts, false)
	tokMust(CRLCLTOK, ts, " (parsing 'for' statements)")
	return NewForNode(initExpr, guard, incrExpr, body, lineno)
}

// Parses a display statement (it's either followed by an expression or the end of the statement)
func parseDpy(ts *tokenStream, lineno int) AstNode {
	tok := ts.get()
	switch tok.ttype {
	case SCOLTOK, EOFTOK:
		ts.rewind(tok)
		return &DpyNode{NewVarNode("_", lineno), false, false, 0, lineno}
	case COLONTOK:
		tok = ts.get()
		if tok.ttype != SYMTOK {
			unexpectedToken(tok, " (while parsing display statement)")
		}
		switch tok.val {
		case "p":
			return &DpyNode{nil, true, false, 0, lineno}
		case "f":
			CommaMode = floatComma
			return &DpyNode{nil, false, true, floatComma, lineno}
		case "r":
			CommaMode = rationalComma
			return &DpyNode{nil, false, true, rationalComma, lineno}
		default:
			unexpectedToken(tok, " (while parsing display statement)")
		}
	}

	ts.rewind(tok)
	expr := parseExpressionSet(ts)
	return &DpyNode{expr, false, false, 0, lineno}
}

func parseExit(ts *tokenStream, lineno int) AstNode {
	tok := ts.get()
	ts.rewind(tok)
	if tok.ttype != SCOLTOK && tok.ttype != EOFTOK {
		unexpectedToken(tok, " (while parsing exit statement)")
	}
	return &ExitNode{lineno}
}

// Parses assignment expression, this only does the infix operator parsing, everything else is offloaded to parseExpressionNoinfix
// expressionSet ::= <var> <assignment-operator> <expressionInfix> | <expressionInfix>
func parseExpressionSet(ts *tokenStream) AstNode {
	tok1 := ts.get()
	if tok1.ttype != SYMTOK {
		ts.rewind(tok1)
		return parseExpressionInfix(ts)
	}

	tok2 := ts.get() // fun fact: this the thing that makes this grammar LL(2) instead of LL(1)
	if tok2.ttype.IsSetOperator {
		return NewSetOpNode(tok2, tok1.val, parseExpressionInfix(ts))
	}

	ts.rewind(tok2)
	ts.rewind(tok1)
	return parseExpressionInfix(ts)
}

// Parses infix expression using dijkstra algorithm:
// https://en.wikipedia.org/wiki/Shunting-yard_algorithm
//
// Equivalent BNF productions:
// expressionComp ::= <expressionBool> <comparison-operator> <expressionComp> | <expressionBool>
// expressionBool ::= <expressionAdd> <bool-opeartor> <expressionBool> | <expressionAdd>
// expressionAdd ::= <expressionMul> <add-or-subtract> <expressionAdd> | <expressionMul>
// expressionMul ::= <expressionNoninfix> <mul-or-div> <expressionMul> | <expressionNoninfix>
func parseExpressionInfix(ts *tokenStream) AstNode {
	outStack := []AstNode{}
	opStack := []token{}

	outpop := func() AstNode {
		r := outStack[len(outStack)-1]
		outStack = outStack[:len(outStack)-1]
		return r
	}

	oppop := func() {
		op := opStack[len(opStack)-1]
		opStack = opStack[:len(opStack)-1]
		right := outpop()
		left := outpop()
		outStack = append(outStack, NewBinOpNode(op, left, right))
	}

	for {
		outStack = append(outStack, parseExpressionNoinfix(ts))

		tokop := ts.get()
		if tokop.ttype == EOFTOK || tokop.ttype == PARCLTOK || tokop.ttype == SCOLTOK || tokop.ttype == COMMATOK {
			ts.rewind(tokop)
			break
		}

		if tokop.ttype.Priority < 0 {
			unexpectedToken(tokop, "expecting operator")
		}

		for {
			if len(opStack) == 0 {
				break
			}

			prevop := opStack[len(opStack)-1]
			if tokop.ttype.Priority > prevop.ttype.Priority {
				break
			}
			// if tokop is right associative break when tokop.ttype.Priority >= prevop.ttype.Priority
			oppop()
		}

		opStack = append(opStack, tokop)
	}

	for len(opStack) > 0 {
		oppop()
	}

	if len(outStack) != 1 {
		panic(fmt.Errorf("could not parse"))
	}

	return outStack[0]
}

// Parses everything related to expressions except infix operators (because they are hard)
// expressionNoinfix ::= <literal> | <symbol>++ | <symbol>-- | <symbol>(<expression>, …) | <symbol> | +<expressionNoinfix> | -<expressionNoinfix> | !<expressionNoinfix> | (<expression>)
func parseExpressionNoinfix(ts *tokenStream) AstNode {
	tok := ts.get()

	switch tok.ttype {
	/* leaves */
	case REALTOK:
		return parseReal(tok.val, tok.lineno)
	case INTTOK:
		return parseInt(tok.val, 10, tok.lineno)
	case HEXTOK:
		return parseInt(tok.val[2:], 16, tok.lineno)
	case OCTTOK:
		return parseInt(tok.val[1:], 8, tok.lineno)
	case DATETOK:
		return parseDate(tok.val, tok.lineno)
	case TIMETOK:
		return parseTime(tok.val, tok.lineno)

	/* variables, function calls, postfix operators */
	case SYMTOK:
		tok2 := ts.get()
		switch tok2.ttype {

		/* variable with postifix operator */
		case INCOPTOK:
			fallthrough
		case DECOPTOK:
			return NewUniOpNode(tok2, NewVarNode(tok.val, tok.lineno))

		/* function call */
		case PAROPTOK:
			return parseFnCall(tok.val, ts, tok.lineno)

		/* just a simple variable */
		default:
			ts.rewind(tok2)
			return NewVarNode(tok.val, tok.lineno)
		}

	/* prefix unary operators */
	case ADDOPTOK:
		return parseExpressionNoinfix(ts)
	case SUBOPTOK:
		return NewUniOpNode(tok, parseExpressionNoinfix(ts))
	case NEGOPTOK:
		return NewUniOpNode(tok, parseExpressionNoinfix(ts))

	/* subexpression */
	case PAROPTOK:
		n := parseExpressionSet(ts)
		tokMust(PARCLTOK, ts, " (while parsing subexpression)")
		return n

	}

	unexpectedToken(tok, " (while parsing basic expression)")
	panic("Unreachable")
}

// parses a function call, both the name of the function and the parenthesis have already been parsed
func parseFnCall(name string, ts *tokenStream, lineno int) AstNode {
	args := []AstNode{}
	first := true
	for {
		tok := ts.get()
		if tok.ttype == PARCLTOK {
			return NewFnCallNode(name, args, lineno)
		}

		if !first {
			if tok.ttype != COMMATOK {
				unexpectedToken(tok, " (while parsing function call)")
			}
		} else {
			ts.rewind(tok)
			first = false
		}

		args = append(args, parseExpressionSet(ts))
	}

	panic("Unreachable")
}

// Parses a real number
func parseReal(s string, lineno int) AstNode {
	switch CommaMode {
	case undefinedComma:
		panic(fmt.Errorf("Can not parse numbers with a comma in undefined mode, use '@:f' for floating point or '@:r' for rational"))
	case floatComma:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(fmt.Errorf("Syntax error: wrong number format at line %d: %s", lineno, err.Error()))
		}
		flavor := DECFLV
		if strings.Index(s, "e") >= 0 || strings.Index(s, "E") >= 0 {
			flavor = EXPFLV
		}
		return NewConstNode(newFloatval(v, flavor), lineno)
	case rationalComma:
		var v big.Rat
		_, ok := v.SetString(s)
		if !ok {
			panic(fmt.Errorf("Syntax error: wrong number format at line %d", lineno))
		}
		return NewConstNode(newRatval(v, max(1, strprec(s))), lineno)
	default:
		panic("unexpected")
	}
}

// if s is a string representing a floating point number it returns the
// number of digits after the comma.
func strprec(s string) int {
	dot := strings.Index(s, ".")
	if dot < 0 {
		return 0
	}
	for i := dot + 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return i - dot - 1
		}
	}
	return len(s) - dot - 1
}

// Parses an integer
func parseInt(s string, base, lineno int) AstNode {
	var v big.Int
	_, ok := v.SetString(s, base)
	if !ok {
		panic(fmt.Errorf("Syntax error: wrong number format at line %d", lineno))
	}
	flavor := DECFLV
	switch base {
	case 16:
		flavor = HEXFLV
	case 8:
		flavor = OCTFLV
	}
	return NewConstNode(newIntval(v, flavor), lineno)
}

// Parses a date
func parseDate(s string, lineno int) AstNode {
	t, err := time.Parse("20060102", s)
	if err != nil {
		panic(fmt.Errorf("Syntax error: wrong date format at line: %d: %v", lineno, err.Error()))
	}
	return NewDateNode(t, lineno)
}

// Parses a time
func parseTime(s string, lineno int) AstNode {
	v := strings.Split(s, ":")
	r := int64(0)
	for _, x := range v {
		xv, err := strconv.Atoi(x)
		if err != nil {
			panic(fmt.Errorf("Syntax error: wrong time format at line: %d: %v", lineno, err.Error()))
		}
		r *= 60
		r += int64(xv)
	}
	return NewTimeNode(r, lineno)
}

// Extracts a token from the stream, checks that it's the given type and returns its value
// if there are no more tokens or the wrong token is found panics
func tokMust(ttype tokenType, ts *tokenStream, when string) string {
	tok := ts.get()

	if tok.ttype == ERRTOK {
		panic(fmt.Errorf("%s", tok.val))
	} else if tok.ttype == EOFTOK {
		panic(fmt.Errorf("Syntax error: unexpected end of file at line %d %s", tok.lineno, when))
	} else if tok.ttype != ttype {
		panic(fmt.Errorf("Syntax error: unexpected token '%s' in line %d (expecting '%v' %s)", tok.val, tok.lineno, ttype, when))
	}

	return tok.val
}

func unexpectedToken(tok token, when string) {
	if tok.ttype == ERRTOK {
		panic(fmt.Errorf("%s %s", tok.val, when))
	} else if tok.ttype == EOFTOK {
		panic(fmt.Errorf("Syntax error: unexpected end of file at line %d %s", tok.lineno, when))
	}
	panic(fmt.Errorf("Syntax error: unexpected token '%s' in line %d %s", tok.val, tok.lineno, when))
}
