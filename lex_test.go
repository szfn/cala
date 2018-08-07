package main

import (
	"strings"
	"testing"
)

func tokEqual(t *testing.T, tt []token, te []token) {
	if len(tt) != len(te) {
		t.Fatalf("Token number mismatch:\ngot:%v\nexp:%v\n", tt, te)
	}

	for i, _ := range tt {
		if tt[i].ttype != te[i].ttype {
			t.Fatalf("Type mismatch at token %d: %v %v\n", i, tt[i], te[i])
		}

		if tt[i].val != te[i].val {
			t.Fatalf("Value mismatch at token %d: %s %s\n", i, tt[i].val, te[i].val)
		}

		if tt[i].lineno != te[i].lineno {
			t.Fatalf("Line number mismatch at token %d: %v %v\n", i, tt[i], te[i])
		}
	}
}

func TestLexerOk(t *testing.T) {
	s := `
# This is a comment, it should be ignored
0123	0x1aAf5 0 45
.345  	 0.345 3.345 	32.3126e2  32.3126e-2 32.3126e+2
  _something something som1231 12'345.2 12'345
`

	expected := []token{
		{OCTTOK, "0123", 3},
		{HEXTOK, "0x1aAf5", 3},
		{INTTOK, "0", 3},
		{INTTOK, "45", 3},

		{REALTOK, ".345", 4},
		{REALTOK, "0.345", 4},
		{REALTOK, "3.345", 4},
		{REALTOK, "32.3126e2", 4},
		{REALTOK, "32.3126e-2", 4},
		{REALTOK, "32.3126e+2", 4},

		{SYMTOK, "_something", 5},
		{SYMTOK, "something", 5},
		{SYMTOK, "som1231", 5},

		{REALTOK, "12345.2", 5},
		{INTTOK, "12345", 5},

		{EOFTOK, "", 6},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerFail(t *testing.T) {
	s := `
# This is a comment, it should be ignored
0123	01238
`

	expected := []token{
		{OCTTOK, "0123", 3},
		{OCTTOK, "0123", 3}, // ← spurious token emitted here
		{ERRTOK, "Syntax error: unexpected character '8' in line 3", 3},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerOps(t *testing.T) {
	s := `
a+b**z(//=32.2)
`

	expected := []token{
		{SYMTOK, "a", 2},
		{ADDOPTOK, "+", 2},
		{SYMTOK, "b", 2},
		{POWOPTOK, "**", 2},
		{SYMTOK, "z", 2},
		{PAROPTOK, "(", 2},
		{DIVOPTOK, "/", 2},
		{DIVEQTOK, "/=", 2},
		{REALTOK, "32.2", 2},
		{PARCLTOK, ")", 2},
		{EOFTOK, "", 3},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerOk2(t *testing.T) {
	s := "2+2"
	expected := []token{
		{INTTOK, "2", 1},
		{ADDOPTOK, "+", 1},
		{INTTOK, "2", 1},
		{EOFTOK, "", 1},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerOk3(t *testing.T) {
	s := "2**(1/2)"
	expected := []token{
		{INTTOK, "2", 1},
		{POWOPTOK, "**", 1},
		{PAROPTOK, "(", 1},
		{INTTOK, "1", 1},
		{DIVOPTOK, "/", 1},
		{INTTOK, "2", 1},
		{PARCLTOK, ")", 1},
		{EOFTOK, "", 1},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerOk4(t *testing.T) {
	s := "2 %= 2"
	expected := []token{
		{INTTOK, "2", 1},
		{MODEQTOK, "%=", 1},
		{INTTOK, "2", 1},
		{EOFTOK, "", 1},
	}

	tokens := lexAll(strings.NewReader(s))

	tokEqual(t, tokens, expected)
}

func TestLexerOk5(t *testing.T) {
	s := "!x++"

	expected := []token{
		{NEGOPTOK, "!", 1},
		{SYMTOK, "x", 1},
		{INCOPTOK, "++", 1},
		{EOFTOK, "", 1},
	}

	tokens := lexAll(strings.NewReader(s))
	tokEqual(t, tokens, expected)
}

func TestDateToks(t *testing.T) {
	s := "$20150101"
	expected := []token{
		{DATETOK, "20150101", 1},
		{EOFTOK, "", 1},
	}

	tokens := lexAll(strings.NewReader(s))
	tokEqual(t, tokens, expected)
}
