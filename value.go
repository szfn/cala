package main

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
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
	prec   int
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
	EXPFLV
	TIMEFLV
)

func newZeroVal(kind valueKind, flavor valueFlavor, prec int) *value {
	return &value{kind, flavor, big.Int{}, 0, big.Rat{}, nil, nil, nil, prec}
}

func newDateval(t time.Time) *value {
	return &value{DTVAL, DECFLV, big.Int{}, 0, big.Rat{}, nil, &t, nil, 0}
}

func newFloatval(x float64, flavor valueFlavor) *value {
	return &value{DVAL, flavor, big.Int{}, x, big.Rat{}, nil, nil, nil, 0}
}

func newFloatvalDerived(x float64, a1, a2 *value) *value {
	flavor := DECFLV
	if a1.flavor == EXPFLV || a2.flavor == EXPFLV {
		flavor = EXPFLV
	}
	return newFloatval(x, flavor)
}

func newRatval(v big.Rat, prec int) *value {
	return &value{RVAL, DECFLV, big.Int{}, 0, v, nil, nil, nil, prec}
}

func newIntval(v big.Int, flavor valueFlavor) *value {
	return &value{IVAL, flavor, v, 0, big.Rat{}, nil, nil, nil, 0}
}

func newBoolval(b bool) *value {
	v := newZeroVal(IVAL, DECFLV, 0)
	if b {
		v.ival = *big.NewInt(1)
	}
	return v
}

func makeFuncValue(nargs int, fn BuiltinFunc) *value {
	return &value{BVAL, DECFLV, big.Int{}, 0, big.Rat{}, nil, nil, &BuiltinFn{nargs: nargs, fn: fn}, 0}
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
		case TIMEFLV:
			var h, m, s, t big.Int
			h.Div(&vv.ival, big.NewInt(60*60))
			t.Mul(&h, big.NewInt(60*60))
			m.Sub(&vv.ival, &t)
			s.Set(&m)
			m.Div(&m, big.NewInt(60))
			t.Mul(&m, big.NewInt(60))
			s.Sub(&s, &t)
			if h.Cmp(big.NewInt(0)) > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", &h, &m, &s)
			}
			return fmt.Sprintf("%02d:%02d", &m, &s)
		default:
			if programmerMode {

				return fmt.Sprintf("%s\t%#x", fmtfloatstr(vv.ival.String()), &vv.ival)
			} else {
				return fmtfloatstr(vv.ival.String())
			}
		}
	case DVAL:
		if vv.flavor == EXPFLV {
			return fmtfloatstr(strconv.FormatFloat(vv.dval, 'g', -1, 64))
		} else {
			return fmtfloatstr(strconv.FormatFloat(vv.dval, 'f', -1, 64))
		}
	case RVAL:
		return fmtfloatstr(vv.rval.FloatString(vv.prec))
	case DTVAL:
		return "$" + vv.dtval.Format("20060102")
	}
	return fmt.Sprintf("@")
}

func max(v ...int) int {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func min(v ...int) int {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

func fmtfloatstr(s string) string {
	if strings.Index(s, "e") >= 0 || strings.Index(s, "E") >= 0 {
		return s
	}

	sign := ""
	integral := s
	frac := ""

	dot := strings.Index(s, ".")
	if dot >= 0 {
		integral = s[:dot]
		frac = s[dot:]
	}
	if len(integral) > 0 && (integral[0] < '0' || integral[0] > '9') {
		sign = integral[:1]
		integral = integral[1:]
	}

	found := false
	for i := len(frac) - 1; i >= 1; i-- {
		if frac[i] != '0' {
			frac = frac[:i+1]
			found = true
			break
		}
	}
	if !found {
		frac = frac[:min(2, len(frac))]
	}

	newintegral := make([]byte, 0, len(integral)+len(integral)/3)

	for i := 0; i < len(integral); i++ {
		j := len(integral) - i
		if j%3 == 0 && i != 0 && j != 0 {
			newintegral = append(newintegral, '\'')
		}
		newintegral = append(newintegral, integral[i])
	}

	return sign + string(newintegral) + frac
}
