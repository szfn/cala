package main

import (
	"fmt"
	"math"
	"testing"
)

func execString(t *testing.T, s string) *value {
	pgm, err := parseString(s)
	if err != nil {
		fmt.Printf("Program:\n%s\n", s)
		fmt.Printf("%v\n", err)
		t.Fatalf("")
	}
	return pgm.Exec(NewCallStack())
}

func testExecReal(t *testing.T, s string, tgt float64) {
	v := execString(t, s)
	if v.kind != DVAL {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value not real: %v\n", v)
	}
	if math.Abs(v.dval - tgt) > 0.0001 {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value mismatch %v (expected: %g)\n", v, tgt)
	}
}

func testExecInt(t *testing.T, s string, tgt int64) {
	v := execString(t, s)
	if v.kind != IVAL {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value not integer: %v\n", v)
	}
	if v.ival != tgt {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value mismatch %v (expected: %g)\n", v, tgt)
	}
}

func TestExecOps(t *testing.T) {
	// reals
	testExecReal(t, "10.2", 10.2)
	testExecReal(t, "-10.2", -10.2)
	testExecReal(t, "+10.2", 10.2)
	testExecReal(t, "2.2 + 2.1", 4.3)
	testExecReal(t, "2.1 - 0.5", 1.6)
	testExecReal(t, "2.0 * 5.0", 10.0)
	testExecReal(t, "5/2", 2.5)
	testExecReal(t, "2**4", 16)
	testExecReal(t, "2**-2", 0.25)
	testExecReal(t, "2**0.5", 1.41421356)
	testExecReal(t, "2**(1/2)", 1.41421356)
	testExecInt(t, "10.3 == 11.3", 0)
	testExecInt(t, "11.3 == 11.3", 1)
	testExecInt(t, "10.3 != 11.3", 1)
	testExecInt(t, "11.3 != 11.3", 0)
	testExecInt(t, "10.2 > 11.1", 0)
	testExecInt(t, "11.5 > 10.9", 1)
	testExecInt(t, "10.2 >= 11.1", 0)
	testExecInt(t, "11.5 >= 10.9", 1)
	testExecInt(t, "10.2 < 11.1", 1)
	testExecInt(t, "11.5 < 10.9", 0)
	testExecInt(t, "10.2 <= 11.1", 1)
	testExecInt(t, "11.5 <= 10.9", 0)

	// integers
	testExecInt(t, "2", 2)
	testExecInt(t, "-2", -2)
	testExecInt(t, "!0", 1)
	testExecInt(t, "!1", 0)
	testExecInt(t, "+2", 2)
	testExecInt(t, "2 + 2", 4)
	testExecInt(t, "2 - 5", -3)
	testExecInt(t, "2 * 5", 10)
	testExecInt(t, "4 % 3", 1)
	testExecInt(t, "1 || 0", 1)
	testExecInt(t, "0 || 0", 0)
	testExecInt(t, "1 && 0", 0)
	testExecInt(t, "1 && 1", 1)
	testExecInt(t, "10 == 11", 0)
	testExecInt(t, "11 == 11", 1)
	testExecInt(t, "10 != 11", 1)
	testExecInt(t, "11 != 11", 0)
	testExecInt(t, "10 > 11", 0)
	testExecInt(t, "11 > 10", 1)
	testExecInt(t, "10 >= 11", 0)
	testExecInt(t, "11 >= 10", 1)
	testExecInt(t, "10 < 11", 1)
	testExecInt(t, "11 < 10", 0)
	testExecInt(t, "10 <= 11", 1)
	testExecInt(t, "11 <= 10", 0)
}

func TestExecVars(t *testing.T) {
	testExecInt(t, "a = 12; a++; a", 13)
	testExecInt(t, "a = 12; a--; a", 11)
	testExecInt(t, "a = 12; a++", 13)
	testExecInt(t, "a = 12; a--", 11)
	testExecReal(t, "a = 1.2; a + 2", 3.2)

	// test real
	testExecReal(t, "a = 2.3; a += 2; a", 4.3)
	testExecReal(t, "a = 2.3; a -= 2; a", 0.3)
	testExecReal(t, "a = 2.3; a *= 2; a", 4.6)
	testExecReal(t, "a = 2.3; a /= 2; a", 1.15)
	testExecReal(t, "a = 2.3; a += 2", 4.3)
	testExecReal(t, "a = 2.3; a -= 2", 0.3)
	testExecReal(t, "a = 2.3; a *= 2", 4.6)
	testExecReal(t, "a = 2.3; a /= 2", 1.15)

	// test integer
	testExecInt(t, "a = 2; a += 2; a", 4)
	testExecInt(t, "a = 2; a -= 2; a", 0)
	testExecInt(t, "a = 2; a *= 2; a", 4)
	testExecReal(t, "a = 2; a /= 2; a", 1.0)
	testExecInt(t, "a = 3; a %= 2; a", 1)
	testExecInt(t, "a = 2; a += 2", 4)
	testExecInt(t, "a = 2; a -= 2", 0)
	testExecInt(t, "a = 2; a *= 2", 4)
	testExecReal(t, "a = 2; a /= 2", 1.0)
	testExecInt(t, "a = 3; a %= 2", 1)
}

func TestExecComposed(t *testing.T) {
	testExecInt(t, "1+1+1+1+1", 5)
	testExecInt(t, "1+2*3-2", 5)
	testExecInt(t, "(1+2)*3-2", 7)
	testExecInt(t, "a = 12; b = 1 + !a++; b + a", 14)
	testExecReal(t, "(2 + 3) / ( 2 * 3)", 5.0/6.0)
}

func TestExecStatement(t *testing.T) {
	testExecInt(t, `
		a = 2;
		if (a > 1) {
			b = 1;
		} else {
			b = 2;
		}
		b`, 1)
	testExecInt(t, `
		a = -2;
		if (a > 1) {
			b = 1;
		} else {
			b = 2;
		}
		b`, 2)
	testExecInt(t, `
		a = 5;
		b = 2;
		while (a > 0) {
			a--;
			b *= 2;
		}
		b`, 64)
	testExecInt(t, `
		b = 2;
		for(i = 0; i < 5; i++) {
			b *= 2;
		}
		b`, 64)
	testExecInt(t, `
		b = 0;
		func addfn(a) {
			b += a;
		}
		for (i = 0; i < 5; i++) {
			addfn(i);
		}
		b`, 10)
}

func TestBuiltin(t *testing.T) {
	testExecReal(t, "cos(3)", math.Cos(3))
	testExecReal(t, "ln(4.3)", math.Log(4.3))
}

