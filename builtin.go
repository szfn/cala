package main

import (
	"fmt"
	"math"
	"strings"
)

func intAbs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

var btnAbs = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		switch argv[0].kind {
		case IVAL:
			return &value{IVAL, intAbs(argv[0].ival), 0, nil, nil}
		case DVAL:
			return &value{DVAL, 0, math.Abs(argv[0].dval), nil, nil}
		}
		panic(fmt.Errorf("Can not apply abs to non-number value"))
	},
}}

var btnAcos = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Acos(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnAsin = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Asin(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnAtan = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Atan(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnCos = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Cos(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnCosh = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Cosh(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnFloor = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{IVAL, int64(math.Floor(argv[0].Real(lineno))), 0, nil, nil}
	},
}}

var btnCeil = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{IVAL, int64(math.Ceil(argv[0].Real(lineno))), 0, nil, nil}
	},
}}

var btnLn = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnLog10 = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log10(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnLog2 = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log2(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnSin = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sin(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnSinh = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sinh(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnSqrt = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sqrt(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnTan = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Tan(argv[0].Real(lineno)), nil, nil}
	},
}}

var btnTanh = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Tanh(argv[0].Real(lineno)), nil, nil}
	},
}}

func hexsplit(s string) string {
	r := []string{}
	for i := 0; i < len(s); i += 4 {
		r = append(r, s[i:i+4])
	}
	return strings.Join(r, " ")
}

func binaryPrint(x uint64) {
	for i := 0; i < 64; i++ {
		bit := (x & 0x8000000000000000) >> 63
		x <<= 1
		fmt.Printf("%d", bit)
		if (i+1)%4 == 0 {
			fmt.Printf(" ")
		}
		if (i+1)%16 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

var btnDpy = &value{BVAL, 0, 0, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		var x uint64
		switch argv[0].kind {
		case IVAL:
			fmt.Printf("dec = %d\n", argv[0].ival)
			x = uint64(argv[0].ival)
		case DVAL:
			fmt.Printf("dec = %g\n", argv[0].dval)
			x = math.Float64bits(argv[0].dval)
		default:
			fmt.Printf("not a number\n")
		}

		fmt.Printf("oct = %o\n", x)
		fmt.Printf("hex = %s\n", hexsplit(fmt.Sprintf("%016X", x)))

		binaryPrint(x)

		return &value{IVAL, 0, 0, nil, nil}
	},
}}
