package main

import (
	"strings"
	"math"
	"fmt"
)

var TokenTypes = map[string]tokenType{ }
var OpPriority = [][]tokenType{}

type tokenType *tokenTypeDef
type tokenTypeDef struct {
	Token *tokenTypeDef
	Name string // printable name of the token
	XName string // exact parsing name

	IsSetOperator bool // is set operator (=, +=, -=, etc)

	Priority int // operator priority, set to -1 for non-operators
	BinFn func(*value, *value, int) *value // for binary operators this is the function called on execution
	UniFn func(*value, int) *value // for unary operators this is the function called on execution

	LexFollow []tokenType // lexing aid, if this operator is also the prefix for other tokens put the other tokens here
}

func registerTokenType(t tokenType) {
	TokenTypes[t.XName] = t

	/* register to OpPriority table */
	if t.Priority != -1 {
		if t.Priority >= len(OpPriority) {
			NewOpPriority := make([][]tokenType, t.Priority+1, t.Priority+1)
			copy(NewOpPriority, OpPriority)
			OpPriority = NewOpPriority
		}
		if OpPriority[t.Priority] == nil {
			OpPriority[t.Priority] = []tokenType{}
		}
		OpPriority[t.Priority] = append(OpPriority[t.Priority], t)
	}

	/* Automatically compte lexer followers */
	lexFollow := []tokenType{}
	for _, tt := range TokenTypes {
		if strings.HasPrefix(tt.XName, t.XName) {
			lexFollow = append(lexFollow, tt)
		}

		if strings.HasPrefix(t.XName, tt.XName) {
			//println("Adding " + t.XName + " to followers of " + tt.XName)
			if tt.LexFollow == nil {
				tt.LexFollow = []tokenType{ t }
			} else {
				tt.LexFollow = append(tt.LexFollow, t)
			}
		}
	}

	if len(lexFollow) > 0 {
		//print("Setting " + t.XName + " followers to:")
		//for _, tt := range lexFollow {
		//	print(" ", tt.XName)
		//}
		//println()
		t.LexFollow = lexFollow
	}
}

func T(name string) tokenType {
	r := &tokenTypeDef{ nil, name, name, false, -1, nil, nil, nil }
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp2(name string, priority int, binFn func(*value, *value, int) *value) tokenType {
	return TOp2X(name, name, priority, binFn)
}

func TOp(name string, priority int, uniFn func(*value, int) *value) tokenType {
	r := &tokenTypeDef{ nil, name, name, false, priority, nil, uniFn, nil }
	r.Token = r
	registerTokenType(r)
	return r
}

func TSetOp(name string, binFn func(*value, *value, int) *value) tokenType {
	r := &tokenTypeDef{ nil, name, name, true, -1, binFn, nil, nil }
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp12(name string, priority int, binFn func(*value, *value, int) *value, uniFn func(*value, int) *value) tokenType {
	r := &tokenTypeDef{ nil, name, name, false, priority, binFn, uniFn, nil }
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp2X(name, xname string, priority int, binFn func(*value, *value, int) *value) tokenType {
	r := &tokenTypeDef{ nil, name, xname, false, priority, binFn, nil, nil }
	r.Token = r
	registerTokenType(r)
	return r
}

var ERRTOK = T("an error occoured")
var EOFTOK = T("end of file")

var REALTOK = T("a real number")
var INTTOK = T("an integer number")
var HEXTOK = T("a hexadecimal number")
var OCTTOK = T("an octal number")
var KWDTOK = T("a keyword")
var SYMTOK = T("any symbol")

var PAROPTOK = T("(")
var PARCLTOK = T(")")
var CRLOPTOK = T("{")
var CRLCLTOK  = T("}")

var ADDOPTOK = TOp2("+", 2, func(a1, a2 *value, lineno int) *value {
	if (a1.kind == IVAL) && (a2.kind == IVAL) {
		return &value{IVAL, a1.Int(lineno) + a2.Int(lineno), 0, nil, nil}
	}
	return &value{DVAL, 0, a1.Real(lineno) + a2.Real(lineno), nil, nil}
})

var SUBOPTOK = TOp12("-", 2, func(a1, a2 *value, lineno int) *value {
	if (a1.kind == IVAL) && (a2.kind == IVAL) {
		return &value{IVAL, a1.Int(lineno) - a2.Int(lineno), 0, nil, nil}
	}
	return &value{DVAL, 0, a1.Real(lineno) - a2.Real(lineno), nil, nil}
},
func (a1 *value, lineno int) *value {
	switch a1.kind {
	case IVAL:
		return &value{ IVAL, -a1.ival, 0, nil, nil }
	case DVAL:
		return &value{ DVAL, 0, -a1.dval, nil, nil }
	}
	panic(fmt.Errorf("Can not apply operator to non-numer value at line %d", lineno))
})

var MULOPTOK = TOp2("*", 3, func(a1, a2 *value, lineno int) *value {
	if (a1.kind == IVAL) && (a2.kind == IVAL) {
		return &value{IVAL, a1.Int(lineno) * a2.Int(lineno), 0, nil, nil}
	}
	return &value{DVAL, 0, a1.Real(lineno) * a2.Real(lineno), nil, nil}
})

var DIVOPTOK = TOp2("/", 4, func(a1, a2 *value, lineno int) *value {
	return &value{DVAL, 0, a1.Real(lineno) / a2.Real(lineno), nil, nil}
})

var MODOPTOK = TOp2("%", 4, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, a1.Int(lineno) % a2.Int(lineno), 0, nil, nil}
})

var POWOPTOK = TOp2("**", 4, func(a1, a2 *value, lineno int) *value {
	return &value{DVAL, 0, math.Pow(a1.Real(lineno), a2.Real(lineno)), nil, nil}
})

var OROPTOK = TOp2("||", 1, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Bool(lineno) || a2.Bool(lineno)), 0, nil, nil}
})

var BWOROPTOK = TOp2("|", 1, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, a1.Int(lineno) | a2.Int(lineno), 0, nil, nil}
})

var ANDOPTOK = TOp2("&&", 1, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Bool(lineno) && a2.Bool(lineno)), 0, nil, nil}
})

var BWANDOPTOK = TOp2("&", 1, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, a1.Int(lineno) & a2.Int(lineno), 0, nil, nil}
})



var INCOPTOK = TOp("++", -1, func(a1 *value, lineno int) *value {
	switch a1.kind {
	case IVAL:
		a1.ival++
	case DVAL:
		a1.dval++
	default:
		panic(fmt.Errorf("Can not increment function variable at line %d", lineno))
	}
	vv := *a1
	return &vv
})
var DECOPTOK = TOp("--", -1, func(a1 *value, lineno int) *value {
	switch a1.kind {
	case IVAL:
		a1.ival--
	case DVAL:
		a1.dval--
	default:
		panic(fmt.Errorf("Can not decrement function variable at line %d", lineno))
	}
	vv := *a1
	return &vv
})

var NEGOPTOK = TOp("!", -1, func(a1 *value, lineno int) *value {
	if a1.kind != IVAL {
		panic(fmt.Errorf("Can not negate non-integer value at line %d", lineno))
	}
	return &value{ IVAL, asInt(!(a1.ival != 0)), 0, nil, nil }
})

var EQOPTOK = TOp2("==", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) == a2.Real(lineno)), 0, nil, nil}
})

var GEOPTOK = TOp2X("ge", ">=", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) >= a2.Real(lineno)), 0, nil, nil}
})

var GTOPTOK = TOp2X("gt", ">", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) > a2.Real(lineno)), 0, nil, nil}
})

var LEOPTOK = TOp2X("le", "<=", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) <= a2.Real(lineno)), 0, nil, nil}
})

var LTOPTOK = TOp2X("lt", "<", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) < a2.Real(lineno)), 0, nil, nil}
})

var NEOPTOK = TOp2("!=", 0, func(a1, a2 *value, lineno int) *value {
	return &value{IVAL, asInt(a1.Real(lineno) != a2.Real(lineno)), 0, nil, nil}
})

var SETOPTOK = TSetOp("=", nil)
var ADDEQTOK = TSetOp("+=", ADDOPTOK.BinFn)
var SUBEQTOK = TSetOp("-=", SUBOPTOK.BinFn)
var MULEQTOK = TSetOp("*=", MULOPTOK.BinFn)
var DIVEQTOK = TSetOp("/=", DIVOPTOK.BinFn)
var MODEQTOK = TSetOp("%=", MODOPTOK.BinFn)

var COMMATOK = T(",")
var SCOLTOK = T(";")

var KwdTable = map[string]bool{
	"if": true,
	"else": true,
	"while": true,
	"for": true,
	"func": true,
}
