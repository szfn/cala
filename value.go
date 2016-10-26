package main

import (
	"fmt"
	"math/big"
	"time"
)

type value struct {
	kind  valueKind
	ival  big.Int
	dval  float64
	nval  *FnDefNode
	dtval *time.Time
	bval  *BuiltinFn
}

type valueKind int

const (
	IVAL  valueKind = iota // integer
	DVAL                   // double
	PVAL                   // a subprogram
	BVAL                   // a builtin function
	DTVAL                  // date
)

func newZeroVal(kind valueKind) *value {
	return &value{kind, big.Int{}, 0, nil, nil, nil}
}

func newDateval(t time.Time) *value {
	return &value{DTVAL, big.Int{}, 0, nil, &t, nil}
}

func newFloatval(x float64) *value {
	return &value{DVAL, big.Int{}, x, nil, nil, nil}
}

func newBoolval(b bool) *value {
	v := newZeroVal(IVAL)
	if b {
		v.ival = *big.NewInt(1)
	}
	return v
}

func makeFuncValue(nargs int, fn BuiltinFunc) *value {
	return &value{BVAL, big.Int{}, 0, nil, nil, &BuiltinFn{nargs: nargs, fn: fn}}
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

	return IVAL
}

func (vv *value) String() string {
	switch vv.kind {
	case IVAL:
		return vv.ival.String()
	case DVAL:
		return fmt.Sprintf("%g", vv.dval)
	case DTVAL:
		return vv.dtval.Format("20060102")
	}
	return fmt.Sprintf("@")
}
