package main

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"
)

func intAbs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

var btnAbs = makeFuncValue(1, func(argv []*value, lineno int) *value {
	switch argv[0].kind {
	case IVAL:
		v := newZeroVal(IVAL, argv[0].flavor, 0)
		v.ival.Abs(&argv[0].ival)
		return v
	case RVAL:
		var r big.Rat
		r.Abs(&argv[0].rval)
		return newRatval(r, argv[0].prec)
	case DVAL:
		return newFloatval(math.Abs(argv[0].dval), argv[0].flavor)
	}
	panic(fmt.Errorf("Can not apply abs to non-number value"))
})

func makeFloatFuncValue(fn func(float64) float64) *value {
	return makeFuncValue(1, func(argv []*value, lineno int) *value {
		kind := argv[0].kind
		if kind == IVAL {
			switch CommaMode {
			case undefinedComma:
				panic("real mode undefined, use @:r to select rational or @:f to select floating point")
			case floatComma:
				kind = DVAL
			case rationalComma:
				kind = RVAL
			}
		}
		switch kind {
		case RVAL:
			var r big.Rat
			r.SetFloat64(fn(argv[0].Real(lineno)))
			return newRatval(r, max(12, argv[0].prec))
		default:
			return newFloatval(fn(argv[0].Real(lineno)), argv[0].flavor)
		}
	})
}

var btnAcos = makeFloatFuncValue(math.Acos)
var btnAsin = makeFloatFuncValue(math.Asin)
var btnAtan = makeFloatFuncValue(math.Atan)
var btnCos = makeFloatFuncValue(math.Cos)
var btnCosh = makeFloatFuncValue(math.Cosh)
var btnLn = makeFloatFuncValue(math.Log)
var btnLog10 = makeFloatFuncValue(math.Log10)
var btnLog2 = makeFloatFuncValue(math.Log2)
var btnSin = makeFloatFuncValue(math.Sin)
var btnSinh = makeFloatFuncValue(math.Sinh)
var btnSqrt = makeFloatFuncValue(math.Sqrt)
var btnTan = makeFloatFuncValue(math.Tan)
var btnTanh = makeFloatFuncValue(math.Tanh)

var btnFloor = makeFuncValue(1, func(argv []*value, lineno int) *value {
	switch argv[0].kind {
	case RVAL:
		a := argv[0].Rat(lineno)
		var z, r big.Int
		z.QuoRem(a.Num(), a.Denom(), &r)

		if z.Sign() < 0 {
			if r.Cmp(big.NewInt(0)) != 0 {
				z.Sub(&z, big.NewInt(1))
			}
		}
		return newIntval(z, DECFLV)
	default:
		return newIntval(*big.NewInt(int64(math.Floor(argv[0].Real(lineno)))), DECFLV)
	}
})

var btnCeil = makeFuncValue(1, func(argv []*value, lineno int) *value {
	switch argv[0].kind {
	case RVAL:
		a := argv[0].Rat(lineno)
		var z, r big.Int
		z.QuoRem(a.Num(), a.Denom(), &r)

		if z.Sign() >= 0 {
			if r.Cmp(big.NewInt(0)) != 0 {
				z.Add(&z, big.NewInt(1))
			}
		}
		return newIntval(z, DECFLV)
	default:
		return newIntval(*big.NewInt(int64(math.Ceil(argv[0].Real(lineno)))), DECFLV)
	}
})

func hexsplit(s string) string {
	r := []string{}
	for i := 0; i < len(s); i += 4 {
		r = append(r, s[i:i+4])
	}
	return strings.Join(r, " ")
}

func binaryPrint(x uint64) {
	bc := 6
	fmt.Printf("%d %d: ", bc+1, bc)
	for i := 0; i < 64; i++ {
		bit := (x & 0x8000000000000000) >> 63
		x <<= 1
		fmt.Printf("%d", bit)
		if (i+1)%4 == 0 {
			fmt.Printf(" ")
		}
		if (i+1)%16 == 0 {
			fmt.Printf("\n")
			bc -= 2
			if bc >= 0 {
				fmt.Printf("%d %d: ", bc+1, bc)
			}
		} else if (i+1)%8 == 0 {
			fmt.Printf("| ")
		}
	}
	//fmt.Printf("\n")
}

func bitfield(x uint64) string {
	var buf bytes.Buffer
	first := true

	for i := 0; i < 64; i++ {
		bit := (x & 0x01)
		x >>= 1
		if bit != 0 {
			if !first {
				fmt.Fprintf(&buf, "|")
			}
			first = false
			fmt.Fprintf(&buf, "_BV(%d)", i)
		}
	}

	return buf.String()
}

var btnDpy = makeFuncValue(1, func(argv []*value, lineno int) *value {
	prefixes := []struct {
		mul bool
		f   int
		p   string
	}{
		{false, 1000000000, "G"},
		{false, 1000000, "M"},
		{false, 1000, "k"},
		{true, 1000, "m"},
		{true, 1000000, "μ"},
		{true, 1000000000, "n"},
		{false, 1073741824, "GiB"},
		{false, 1048576, "MiB"},
		{false, 1024, "KiB"},
	}

	prefixprint := func(gt func(mulby int, tgt int) bool, pr func(mylby int, divby int, prefix string)) {
		for _, p := range prefixes {
			if p.mul {
				if gt(p.f, 1) {
					pr(1, p.f, p.p)
					break
				}
			} else {
				if gt(1, p.f) {
					pr(p.f, 1, p.p)
					break
				}
			}
		}
	}

	switch argv[0].kind {
	case IVAL:
		fmt.Printf("integer\n")
		fmt.Printf("dec = %d\n", &argv[0].ival)
		fmt.Printf("oct = %o\n", &argv[0].ival)
		if argv[0].ival.BitLen() <= 64 {
			fmt.Printf("hex = %s\n", hexsplit(fmt.Sprintf("%016X", &argv[0].ival)))
			fmt.Printf("bin =\n")
			binaryPrint(argv[0].ival.Uint64())
			fmt.Printf("bitfield = %s\n", bitfield(argv[0].ival.Uint64()))
		} else {
			fmt.Printf("hex = %X", &argv[0].ival)
		}
		prefixprint(
			func(mulby int, tgt int) bool {
				var x big.Rat
				x.Mul(argv[0].Rat(0), big.NewRat(int64(mulby), 1))
				return x.Cmp(big.NewRat(int64(tgt), 1)) >= 0
			},
			func(divby, mulby int, prefix string) {
				var x big.Rat
				x.Mul(argv[0].Rat(0), big.NewRat(int64(mulby), int64(divby)))
				fmt.Printf("%s%s\n", fmtfloatstr(x.FloatString(3)), prefix)
			})

	case DVAL:
		fmt.Printf("float\n")
		fmt.Printf("dec = %f\n", argv[0].dval)
		fmt.Printf("dec = %e\n", argv[0].dval)
		x := math.Float64bits(argv[0].dval)
		fmt.Printf("hex = %s\n", hexsplit(fmt.Sprintf("%016X", x)))
		fmt.Printf("bin =\n")
		binaryPrint(x)
		prefixprint(
			func(mulby int, tgt int) bool {
				x := argv[0].dval * float64(mulby)
				return x >= float64(tgt)
			},
			func(divby, mulby int, prefix string) {
				fmt.Printf("%f%s\n", argv[0].dval/float64(divby)*float64(mulby), prefix)
			})

	case RVAL:
		fmt.Printf("rational\n")
		fmt.Printf("dec = %s\n", fmtfloatstr(argv[0].rval.FloatString(1023)))
		fmt.Printf("dec = %s\n", argv[0].rval.String())
		prefixprint(
			func(mulby int, tgt int) bool {
				var x big.Rat
				x.Mul(&argv[0].rval, big.NewRat(int64(mulby), 1))
				return x.Cmp(big.NewRat(int64(tgt), 1)) >= 0
			},
			func(divby, mulby int, prefix string) {
				var x big.Rat
				x.Mul(&argv[0].rval, big.NewRat(int64(mulby), int64(divby)))
				fmt.Printf("%s%s\n", fmtfloatstr(x.FloatString(argv[0].prec+3)), prefix)

			})

	case PVAL:
		fmt.Printf("function\n")
		fmt.Printf("%s\n", argv[0].nval.String())

	case BVAL:
		fmt.Printf("builtin\n")

	case DTVAL:
		fmt.Printf("%s\n", argv[0].String())

	default:
		fmt.Printf("not a number\n")
	}

	return newZeroVal(IVAL, DECFLV, 0)
})

var btnPrint = makeFuncValue(1, func(argv []*value, lineno int) *value {
	fmt.Printf("= %s\n\n", argv[0])
	return newZeroVal(IVAL, DECFLV, 0)
})

var btnHelp = makeFuncValue(0, func(argv []*value, lineno int) *value {
	fmt.Printf("Type any expression to calculate the return value\n")
	fmt.Printf("\n")
	fmt.Printf("OPERATORS:\n")
	fmt.Printf("+ - * /\t\tNormal arithmetic operators\n")
	fmt.Printf("%%\t\tModulo\n")
	fmt.Printf("**\t\tPower\n")
	fmt.Printf("|| && !\t\tLogical operators\n")
	fmt.Printf("| &\t\tBitwise logical operators\n")
	fmt.Printf("== != < <= >= >\tComparison operators\n")
	fmt.Printf("var = expr\tAssigns the result of expr to var\n")
	fmt.Printf("var op= expr\tShorthand for var = var op expr\n")
	fmt.Printf("\n")
	fmt.Printf("BUILTIN FUNCTIONS:\n")
	fmt.Printf("abs\tacos\tasin\tatan\n")
	fmt.Printf("cos\tcosh\tfloor\tceil\n")
	fmt.Printf("ln\tlog10\tlog2\tsin\n")
	fmt.Printf("sin\tsinh\tsqrt\ttan\n")
	fmt.Printf("tanh\tdpy\tprint\n")
	fmt.Printf("\n")
	fmt.Printf("@ expr\t\tDetailed variable view, alias for dpy(expr)\n")
	fmt.Printf("@:p\t\tToggles programmer mode (in programmer mode results are shown in decimal and hexadecimal\n")
	fmt.Printf("@:f\t\tToggles float mode (numbers with a comma are interpreted as floating point, division produces a floating point number)\n")
	fmt.Printf("@:r\t\tToggles rational mode (numbers with a comma and division produce exact results)\n")
	fmt.Printf("\n")
	fmt.Printf("DATES:\n")
	fmt.Printf("Date literals are declared with $yyyymmdd for example $20160101 is 2016-01-01, integers can be added to and subtracted from dates.\nTwo date values can also be subtracted.\n")

	return newZeroVal(IVAL, DECFLV, 0)
})
