package main

import (
	"fmt"
	"math/big"
	"time"
)

type value struct {
	kind   valueKind
	flavor valueFlavor
	ival   big.Int
	dval   float64
	rval   big.Rat
	nval   *FnDefNode
	dtval  *time.Time
	bval   *BuiltinFn
}

type valueKind uint8

const (
	IVAL  valueKind = iota // integer
	DVAL                   // double
	RVAL                   // rational number
	PVAL                   // a subprogram
	BVAL                   // a builtin function
	DTVAL                  // date
)

type valueFlavor uint8

const (
	DECFLV valueFlavor = iota
	OCTFLV
	HEXFLV
)

func newZeroVal(kind valueKind, flavor valueFlavor) *value {
	return &value{kind, flavor, big.Int{}, 0, big.Rat{}, nil, nil, nil}
}

func newDateval(t time.Time) *value {
	return &value{DTVAL, DECFLV, big.Int{}, 0, big.Rat{}, nil, &t, nil}
}

func newFloatval(x float64) *value {
	return &value{DVAL, DECFLV, big.Int{}, x, big.Rat{}, nil, nil, nil}
}

func newRatval(v big.Rat) *value {
	return &value{RVAL, DECFLV, big.Int{}, 0, v, nil, nil, nil}
}

func newIntval(v big.Int, flavor valueFlavor) *value {
	return &value{IVAL, flavor, v, 0, big.Rat{}, nil, nil, nil}
}

func newBoolval(b bool) *value {
	v := newZeroVal(IVAL, DECFLV)
	if b {
		v.ival = *big.NewInt(1)
	}
	return v
}

func makeFuncValue(nargs int, fn BuiltinFunc) *value {
	return &value{BVAL, DECFLV, big.Int{}, 0, big.Rat{}, nil, nil, &BuiltinFn{nargs: nargs, fn: fn}}
}

func resultKind(a1, a2 *value) valueKind {
	for _, v := range []*value{a1, a2} {
		for _, kind := range []valueKind{PVAL, BVAL, DTVAL} {
			if v.kind == kind {
				return kind
			}
		}
	}

	if a1.kind == DVAL || a2.kind == DVAL {
		return DVAL
	}

	if a1.kind == RVAL || a2.kind == RVAL {
		return RVAL
	}

	return IVAL
}

func (vv *value) String() string {
	switch vv.kind {
	case IVAL:
		switch vv.flavor {
		case HEXFLV:
			if programmerMode {
				return fmt.Sprintf("%d\t%#x", &vv.ival, &vv.ival)
			} else {
				return fmt.Sprintf("%#x", &vv.ival)
			}
		case OCTFLV:
			return fmt.Sprintf("%#o", &vv.ival)
		default:
			if programmerMode {
				return fmt.Sprintf("%d\t%#x", &vv.ival, &vv.ival)
			} else {
				return vv.ival.String()
			}
		}
	case DVAL:
		return fmt.Sprintf("%g", vv.dval)
	case RVAL:
		return vv.rval.FloatString(1)
	case DTVAL:
		return "$" + vv.dtval.Format("20060102")
	}
	return fmt.Sprintf("@")
}
