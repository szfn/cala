package main

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

var TokenTypes = map[string]tokenType{}

//var OpPriority = [][]tokenType{}

type tokenType *tokenTypeDef
type tokenTypeDef struct {
	Token *tokenTypeDef
	Name  string // printable name of the token
	XName string // exact parsing name

	IsSetOperator bool // is set operator (=, +=, -=, etc)

	Priority int                      // operator priority, set to -1 for non-operators
	BinFn    BinOpFunc                // for binary operators this is the function called on execution
	UniFn    func(*value, int) *value // for unary operators this is the function called on execution

	LexFollow []tokenType // lexing aid, if this operator is also the prefix for other tokens put the other tokens here
}

func registerTokenType(t tokenType) {
	TokenTypes[t.XName] = t

	/* Automatically compte lexer followers */
	lexFollow := []tokenType{}
	for _, tt := range TokenTypes {
		if strings.HasPrefix(tt.XName, t.XName) {
			lexFollow = append(lexFollow, tt)
		}

		if strings.HasPrefix(t.XName, tt.XName) {
			//println("Adding " + t.XName + " to followers of " + tt.XName)
			if tt.LexFollow == nil {
				tt.LexFollow = []tokenType{t}
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
	r := &tokenTypeDef{nil, name, name, false, -1, nil, nil, nil}
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp2(name string, priority int, binFn BinOpFunc) tokenType {
	return TOp2X(name, name, priority, binFn)
}

func TOp(name string, priority int, uniFn func(*value, int) *value) tokenType {
	r := &tokenTypeDef{nil, name, name, false, priority, nil, uniFn, nil}
	r.Token = r
	registerTokenType(r)
	return r
}

func TSetOp(name string, binFn BinOpFunc) tokenType {
	r := &tokenTypeDef{nil, name, name, true, -1, binFn, nil, nil}
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp12(name string, priority int, binFn BinOpFunc, uniFn func(*value, int) *value) tokenType {
	r := &tokenTypeDef{nil, name, name, false, priority, binFn, uniFn, nil}
	r.Token = r
	registerTokenType(r)
	return r
}

func TOp2X(name, xname string, priority int, binFn BinOpFunc) tokenType {
	r := &tokenTypeDef{nil, name, xname, false, priority, binFn, nil, nil}
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
var DATETOK = T("a date constant")

var PAROPTOK = T("(")
var PARCLTOK = T(")")
var CRLOPTOK = T("{")
var CRLCLTOK = T("}")
var DPYSTMTOK = T("@")
var COLONTOK = T(":")

func sortDtval(a1, a2 *value) (b1, b2 *value) {
	if a1.kind == DTVAL {
		return a1, a2
	}
	return a2, a1
}

func badtype(name string, lineno int) error {
	return fmt.Errorf("%d: can not apply %s to non-numeric value", lineno, name)
}

var ADDOPTOK = TOp2("+", 2, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		v := newZeroVal(IVAL, a1.flavor)
		v.ival.Add(a1.Int(lineno), a2.Int(lineno))
		return v
	case DVAL:
		return newFloatval(a1.Real(lineno) + a2.Real(lineno))
	case RVAL:
		v := newZeroVal(RVAL, a1.flavor)
		v.rval.Add(a1.Rat(lineno), a2.Rat(lineno))
		return v
	case DTVAL:
		a1, a2 = sortDtval(a1, a2)
		return newDateval(a1.dtval.AddDate(0, 0, int(a2.Int(lineno).Int64())))
	default:
		panic(badtype("+", lineno))
	}
})

var SUBOPTOK = TOp12("-", 2, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		v := newZeroVal(IVAL, a1.flavor)
		v.ival.Sub(a1.Int(lineno), a2.Int(lineno))
		return v
	case DVAL:
		return newFloatval(a1.Real(lineno) - a2.Real(lineno))
	case RVAL:
		v := newZeroVal(RVAL, a1.flavor)
		v.rval.Sub(a1.Rat(lineno), a2.Rat(lineno))
		return v
	case DTVAL:
		a1, a2 = sortDtval(a1, a2)
		return newDateval(a1.dtval.AddDate(0, 0, -int(a2.Int(lineno).Int64())))
	default:
		panic(badtype("-", lineno))
	}
},
	func(a1 *value, lineno int) *value {
		switch a1.kind {
		case IVAL:
			v := newZeroVal(IVAL, a1.flavor)
			v.ival.Neg(&a1.ival)
			return v
		case DVAL:
			return newFloatval(-a1.dval)
		case RVAL:
			v := newZeroVal(RVAL, a1.flavor)
			v.rval.Neg(&a1.rval)
			return v
		default:
			panic(badtype("-", lineno))
		}
	})

var MULOPTOK = TOp2("*", 3, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		v := newZeroVal(IVAL, a1.flavor)
		v.ival.Mul(a1.Int(lineno), a2.Int(lineno))
		return v
	case DVAL:
		return newFloatval(a1.Real(lineno) * a2.Real(lineno))
	case RVAL:
		v := newZeroVal(RVAL, a1.flavor)
		v.rval.Mul(a1.Rat(lineno), a2.Rat(lineno))
		return v
	default:
		panic(badtype("*", lineno))
	}
})

var DIVOPTOK = TOp2("/", 4, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch CommaMode {
	case undefinedComma:
		panic("Can not use division in undefined mode, use '@:f' for floating point or '@:r' for rational")
	case floatComma:
		return newFloatval(a1.Real(lineno) / a2.Real(lineno))
	case rationalComma:
		r := newZeroVal(RVAL, a1.flavor)
		r.rval.Quo(a1.Rat(lineno), a2.Rat(lineno))
		return r
	default:
		panic("unknown mode")
	}
})

var MODOPTOK = TOp2("%", 4, func(a1, a2 *value, kind valueKind, lineno int) *value {
	v := newZeroVal(IVAL, a1.flavor)
	v.ival.Mod(a1.Int(lineno), a2.Int(lineno))
	return v
})

var POWOPTOK = TOp2("**", 4, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		if a2.Int(lineno).Cmp(&big.Int{}) < 0 {
			switch CommaMode {
			case undefinedComma:
				panic("Can not use division in undefined mode, use '@:f' for floating point or '@:r' for rational")
			case floatComma:
				return newFloatval(math.Pow(a1.Real(lineno), a2.Real(lineno)))
			case rationalComma:
				return rationalPow(a2, a2, lineno)
			default:
				panic("impossible")
			}

		} else {
			v := newZeroVal(IVAL, a1.flavor)
			v.ival.Exp(a1.Int(lineno), a2.Int(lineno), nil)
			return v
		}
	case DVAL:
		return newFloatval(math.Pow(a1.Real(lineno), a2.Real(lineno)))
	case RVAL:
		return rationalPow(a1, a2, lineno)
	default:
		panic(badtype("**", lineno))
	}
})

func rationalPow(a1, a2 *value, lineno int) *value {
	expfr := a2.Rat(lineno)
	if !expfr.IsInt() {
		base, _ := a1.Rat(lineno).Float64()
		exp, _ := a2.Rat(lineno).Float64()
		r := newZeroVal(RVAL, a1.flavor)
		r.rval.SetFloat64(math.Pow(base, exp))
		return r
	}

	exp := expfr.Num()
	swap := false
	if exp.Sign() < 0 {
		swap = true
		exp.Neg(exp)
	}

	base := a1.Rat(lineno)
	numerator := base.Num()
	denominator := base.Denom()

	numerator = numerator.Exp(numerator, exp, nil)
	denominator = denominator.Exp(denominator, exp, nil)

	r := newZeroVal(RVAL, a1.flavor)
	if swap {
		r.rval.SetFrac(denominator, numerator)
	} else {
		r.rval.SetFrac(numerator, denominator)
	}
	return r
}

var OROPTOK = TOp2("||", 1, func(a1, a2 *value, kind valueKind, lineno int) *value {
	return newBoolval(a1.Bool(lineno) || a2.Bool(lineno))
})

var BWOROPTOK = TOp2("|", 1, func(a1, a2 *value, kind valueKind, lineno int) *value {
	v := newZeroVal(IVAL, a1.flavor)
	v.ival.Or(a1.Int(lineno), a2.Int(lineno))
	return v
})

var ANDOPTOK = TOp2("&&", 1, func(a1, a2 *value, kind valueKind, lineno int) *value {
	return newBoolval(a1.Bool(lineno) && a2.Bool(lineno))
})

var BWANDOPTOK = TOp2("&", 1, func(a1, a2 *value, kind valueKind, lineno int) *value {
	v := newZeroVal(IVAL, a1.flavor)
	v.ival.And(a1.Int(lineno), a2.Int(lineno))
	return v
})

var INCOPTOK = TOp("++", -1, func(a1 *value, lineno int) *value {
	switch a1.kind {
	case IVAL:
		a1.ival.Add(&a1.ival, big.NewInt(1))
	case DVAL:
		a1.dval++
	case RVAL:
		a1.rval.Add(&a1.rval, big.NewRat(1, 1))
	default:
		panic(badtype("++", lineno))
	}
	vv := *a1
	return &vv
})
var DECOPTOK = TOp("--", -1, func(a1 *value, lineno int) *value {
	switch a1.kind {
	case IVAL:
		a1.ival.Sub(&a1.ival, big.NewInt(1))
	case DVAL:
		a1.dval--
	case RVAL:
		a1.rval.Sub(&a1.rval, big.NewRat(1, 1))
	default:
		panic(badtype("--", lineno))
	}
	vv := *a1
	return &vv
})

var NEGOPTOK = TOp("!", -1, func(a1 *value, lineno int) *value {
	if a1.kind != IVAL {
		panic(badtype("!", lineno))
	}
	return newBoolval(a1.ival.Cmp(&big.Int{}) == 0)
})

var EQOPTOK = TOp2("==", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) == 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) == a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) == 0)
	default:
		panic(badtype("==", lineno))
	}
})

var GEOPTOK = TOp2X("ge", ">=", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) >= 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) >= a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) >= 0)
	default:
		panic(badtype(">=", lineno))
	}
})

var GTOPTOK = TOp2X("gt", ">", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) > 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) > a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) > 0)
	default:
		panic(badtype(">", lineno))
	}
})

var LEOPTOK = TOp2X("le", "<=", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) <= 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) <= a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) <= 0)
	default:
		panic(badtype("<=", lineno))
	}
})

var LTOPTOK = TOp2X("lt", "<", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) < 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) < a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) < 0)
	default:
		panic(badtype("<", lineno))
	}
})

var NEOPTOK = TOp2("!=", 0, func(a1, a2 *value, kind valueKind, lineno int) *value {
	switch kind {
	case IVAL:
		return newBoolval(a1.Int(lineno).Cmp(a2.Int(lineno)) != 0)
	case DVAL:
		return newBoolval(a1.Real(lineno) != a2.Real(lineno))
	case RVAL:
		return newBoolval(a1.Rat(lineno).Cmp(a2.Rat(lineno)) != 0)
	default:
		panic(badtype("!=", lineno))
	}
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
	"if":    true,
	"else":  true,
	"while": true,
	"for":   true,
	"func":  true,
	"exit":  true,
}
