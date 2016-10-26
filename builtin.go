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
		v := newZeroVal(IVAL, argv[0].flavor)
		v.ival.Abs(&argv[0].ival)
		return v
	case DVAL:
		return newFloatval(math.Abs(argv[0].dval))
	}
	panic(fmt.Errorf("Can not apply abs to non-number value"))
})

func makeFloatFuncValue(fn func(float64) float64) *value {
	return makeFuncValue(1, func(argv []*value, lineno int) *value {
		return newFloatval(fn(argv[0].Real(lineno)))
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
	return &value{IVAL, DECFLV, *big.NewInt(int64(math.Floor(argv[0].Real(lineno)))), 0, nil, nil, nil}
})

var btnCeil = makeFuncValue(1, func(argv []*value, lineno int) *value {
	return &value{IVAL, DECFLV, *big.NewInt(int64(math.Ceil(argv[0].Real(lineno)))), 0, nil, nil, nil}
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

	case DVAL:
		fmt.Printf("float\n")
		fmt.Printf("dec = %g\n", argv[0].dval)
		x := math.Float64bits(argv[0].dval)
		fmt.Printf("hex = %s\n", hexsplit(fmt.Sprintf("%016X", x)))
		fmt.Printf("bin =\n")
		binaryPrint(x)

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

	return newZeroVal(IVAL, DECFLV)
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
	fmt.Printf("tanh\tdpy\n")
	fmt.Printf("\n")
	fmt.Printf("@ expr\t\tDetailed variable view, alias for dpy(expr)\n")
	fmt.Printf("@:p\t\tToggles programmer mode (in programmer mode results are shown in decimal and hexadecimal\n")
	fmt.Printf("\n")
	fmt.Printf("DATES:\n")
	fmt.Printf("Date literals are declared with $yyyymmdd for example $20160101 is 2016-01-01, integers can be added to and subtracted from dates\n")

	//TODO: describe how dates work
	return newZeroVal(IVAL, DECFLV)
})
