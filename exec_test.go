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
		fmt.Printf("Error: %v\n", err)
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
	if math.Abs(v.dval-tgt) > 0.0001 {
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
	if v.ival.Int64() != tgt {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value mismatch %v (expected: %d)\n", v, tgt)
	}
}

func testExecPrint(t *testing.T, s string, tgt string) {
	v := execString(t, s)
	if v.String() != tgt {
		fmt.Printf("Program:\n%s\n", s)
		t.Fatalf("Output value mismatch %q (expected: %q)\n", v.String(), tgt)
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
	testExecReal(t, "2.1**4", 19.4481)
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
	testExecInt(t, "2**4", 16)
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

func TestExecArithmeticBook(t *testing.T) {
	testExecInt(t, "3+2", 5)
	testExecReal(t, "4/3", 4.0/3.0)
	testExecReal(t, "2*(3 + 6/2)/4", 3.0)
	testExecInt(t, "6-9", -3)
	testExecInt(t, "3 + 4 * 2", 11)
	testExecInt(t, "6 + 7 * 8", 62)
	testExecReal(t, "16 / 8 - 2", 0.0)
	testExecReal(t, "3 + 6 * (5 + 4) / 3 - 7", 14.0)
	testExecReal(t, "9 - 5 / (8 - 3) * 2 + 6", 13.0)
	testExecReal(t, "150 / (6 + 3 * 8) - 5", 0.0)
	testExecInt(t, "32 + 3 * 15", 77)
	testExecInt(t, "9 + 6 * (8 - 5)", 27)
	testExecReal(t, " (14 - 5) / (9 - 6)", 3.0)
	testExecReal(t, "5 * 8 + 6 / 6 - 12 * 2", 17.0)
}

func TestExecFlavorPermanence(t *testing.T) {
	testExecPrint(t, "16", "16")
	testExecPrint(t, "0x16", "0x16")
	testExecPrint(t, "011", "011")
	testExecPrint(t, "0x16 + 1", "0x17")
	testExecPrint(t, "1 + 0x16", "23")
	testExecPrint(t, "0x1 + 0x16", "0x17")
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

func TestAckermann(t *testing.T) {
	ackermann := `
		func af(m, n) {
			if (m == 0) {
				n+1;
			} else if ((m > 0) && (n == 0)) {
				af(m-1, 1);
			} else {
				af(m-1, af(m, n-1));
			}
		}
	`

	testExecInt(t, ackermann+"af(0, 0)", 1)
	testExecInt(t, ackermann+"af(0, 1)", 2)
	testExecInt(t, ackermann+"af(0, 2)", 3)
	testExecInt(t, ackermann+"af(0, 3)", 4)
	testExecInt(t, ackermann+"af(0, 4)", 5)

	testExecInt(t, ackermann+"af(1, 0)", 2)
	testExecInt(t, ackermann+"af(1, 1)", 3)
	testExecInt(t, ackermann+"af(1, 2)", 4)
	testExecInt(t, ackermann+"af(1, 3)", 5)
	testExecInt(t, ackermann+"af(1, 4)", 6)

	testExecInt(t, ackermann+"af(2, 0)", 3)
	testExecInt(t, ackermann+"af(2, 1)", 5)
	testExecInt(t, ackermann+"af(2, 2)", 7)
	testExecInt(t, ackermann+"af(2, 3)", 9)
	testExecInt(t, ackermann+"af(2, 4)", 11)

	testExecInt(t, ackermann+"af(3, 0)", 5)
	testExecInt(t, ackermann+"af(3, 1)", 13)
	testExecInt(t, ackermann+"af(3, 2)", 29)
	testExecInt(t, ackermann+"af(3, 3)", 61)
	testExecInt(t, ackermann+"af(3, 4)", 125)
}
