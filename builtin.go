package main

import (
	"bytes"
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

var btnAbs = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		switch argv[0].kind {
		case IVAL:
			return &value{IVAL, intAbs(argv[0].ival), 0, nil, nil, nil}
		case DVAL:
			return &value{DVAL, 0, math.Abs(argv[0].dval), nil, nil, nil}
		}
		panic(fmt.Errorf("Can not apply abs to non-number value"))
	},
}}

var btnAcos = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Acos(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnAsin = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Asin(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnAtan = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Atan(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnCos = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Cos(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnCosh = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Cosh(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnFloor = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{IVAL, int64(math.Floor(argv[0].Real(lineno))), 0, nil, nil, nil}
	},
}}

var btnCeil = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{IVAL, int64(math.Ceil(argv[0].Real(lineno))), 0, nil, nil, nil}
	},
}}

var btnLn = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnLog10 = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log10(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnLog2 = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Log2(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnSin = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sin(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnSinh = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sinh(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnSqrt = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Sqrt(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnTan = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Tan(argv[0].Real(lineno)), nil, nil, nil}
	},
}}

var btnTanh = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		return &value{DVAL, 0, math.Tanh(argv[0].Real(lineno)), nil, nil, nil}
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

var btnDpy = &value{BVAL, 0, 0, nil, nil, &BuiltinFn{
	nargs: 1,
	fn: func(argv []*value, lineno int) *value {
		var x uint64
		switch argv[0].kind {
		case IVAL:
			fmt.Printf("integer\n")
			fmt.Printf("dec = %d\n", argv[0].ival)
			x = uint64(argv[0].ival)
			fmt.Printf("oct = %o\n", x)
			fmt.Printf("hex = %s\n", hexsplit(fmt.Sprintf("%016X", x)))
			fmt.Printf("bin =\n")
			binaryPrint(x)
			fmt.Printf("bitfield = %s\n", bitfield(x))

		case DVAL:
			fmt.Printf("float\n")
			fmt.Printf("dec = %g\n", argv[0].dval)
			x = math.Float64bits(argv[0].dval)
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

		return &value{IVAL, 0, 0, nil, nil, nil}
	},
}}
