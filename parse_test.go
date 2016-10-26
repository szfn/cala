package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func matchAst(t *testing.T, pgm string, expected string) {
	n, err := parse(lex(strings.NewReader(pgm)))
	if err != nil {
		fmt.Printf("Program: %s\n", pgm)
		fmt.Printf("%v\n", err)
		if pe, ok := err.(*ParseError); ok {
			pe.printStackTrace(os.Stdout)
		}
		t.Fatalf("")
	}
	if n.String() != expected {
		fmt.Printf("AST match failed:\nprogram: %s\ngot: %s\nexp: %s\n", pgm, n.String(), expected)
		t.Fatalf("")
	}
}

func TestParseExprNoinfix(t *testing.T) {
	matchAst(t,
		"!x++",
		"BodyNode<[UniOpNode<!, UniOpNode<++, VarNode<x>>>]>")
	matchAst(t,
		"afn(3, !2, b)",
		"BodyNode<[FnCallNode<afn, [ConstNode<0, 3, 0> UniOpNode<!, ConstNode<0, 2, 0>> VarNode<b>]>]>")
	matchAst(t,
		"!afn(a++, a--, (2.3), afn2())",
		"BodyNode<[UniOpNode<!, FnCallNode<afn, [UniOpNode<++, VarNode<a>> UniOpNode<--, VarNode<a>> ConstNode<1, 0, 2.3> FnCallNode<afn2, []>]>>]>")
}

func TestParseExpr(t *testing.T) {
	matchAst(t,
		"2+2",
		"BodyNode<[BinOpNode<+, ConstNode<0, 2, 0>, ConstNode<0, 2, 0>>]>")
	matchAst(t,
		"a += fn3(2+2, 3) * 2",
		"BodyNode<[SetOpNode<+=, a, BinOpNode<*, FnCallNode<fn3, [BinOpNode<+, ConstNode<0, 2, 0>, ConstNode<0, 2, 0>> ConstNode<0, 3, 0>]>, ConstNode<0, 2, 0>>>]>")
	matchAst(t,
		"1 + 2 * 3 - 4",
		"BodyNode<[BinOpNode<-, BinOpNode<+, ConstNode<0, 1, 0>, BinOpNode<*, ConstNode<0, 2, 0>, ConstNode<0, 3, 0>>>, ConstNode<0, 4, 0>>]>")
	matchAst(t,
		"a + b + c * d * e",
		NewBodyNode([]AstNode{
			NewBinOpNode(token{ADDOPTOK, "+", 0},
				NewBinOpNode(token{ADDOPTOK, "+", 0},
					NewVarNode("a", 0),
					NewVarNode("b", 0)),
				NewBinOpNode(token{MULOPTOK, "*", 0},
					NewBinOpNode(token{MULOPTOK, "*", 0},
						NewVarNode("c", 0),
						NewVarNode("d", 0)),
					NewVarNode("e", 0)))}, 0).String())
	matchAst(t,
		"a + b + c * d + e",
		NewBodyNode([]AstNode{
			NewBinOpNode(token{ADDOPTOK, "+", 0},
				NewBinOpNode(token{ADDOPTOK, "+", 0},
					NewBinOpNode(token{ADDOPTOK, "+", 0},
						NewVarNode("a", 0),
						NewVarNode("b", 0)),
					NewBinOpNode(token{MULOPTOK, "*", 0},
						NewVarNode("c", 0),
						NewVarNode("d", 0))),
				NewVarNode("e", 0))}, 0).String())
	matchAst(t,
		"a - b + c",
		NewBodyNode([]AstNode{
			NewBinOpNode(token{ADDOPTOK, "+", 0},
				NewBinOpNode(token{SUBOPTOK, "-", 0},
					NewVarNode("a", 0),
					NewVarNode("b", 0)),
				NewVarNode("c", 0))}, 0).String())
	matchAst(t,
		"a + b - c",
		NewBodyNode([]AstNode{
			NewBinOpNode(token{SUBOPTOK, "-", 0},
				NewBinOpNode(token{ADDOPTOK, "+", 0},
					NewVarNode("a", 0),
					NewVarNode("b", 0)),
				NewVarNode("c", 0))}, 0).String())
}

func TestParseNumber(t *testing.T) {
	matchAst(t, "15", "BodyNode<[ConstNode<0, 15, 0>]>")
	matchAst(t, "0xf", "BodyNode<[ConstNode<0, 15, 0>]>")
	matchAst(t, "017", "BodyNode<[ConstNode<0, 15, 0>]>")
}

func TestParseStatement(t *testing.T) {
	matchAst(t,
		"while(a > 0) { 2 + 2; a--; }",
		"BodyNode<[WhileNode<BinOpNode<gt, VarNode<a>, ConstNode<0, 0, 0>>, BodyNode<[BinOpNode<+, ConstNode<0, 2, 0>, ConstNode<0, 2, 0>> UniOpNode<--, VarNode<a>>]>>]>")
	matchAst(t,
		"if (a == 0) { a = 3; }",
		"BodyNode<[IfNode<BinOpNode<==, VarNode<a>, ConstNode<0, 0, 0>>, BodyNode<[SetOpNode<=, a, ConstNode<0, 3, 0>>]>, nil>]>")
	matchAst(t,
		"if (a == 0) { a = 3; } else { a = 2; }",
		"BodyNode<[IfNode<BinOpNode<==, VarNode<a>, ConstNode<0, 0, 0>>, BodyNode<[SetOpNode<=, a, ConstNode<0, 3, 0>>]>, BodyNode<[SetOpNode<=, a, ConstNode<0, 2, 0>>]>>]>")
	matchAst(t,
		"for (i = 0; i < n; i++) { a = 0; }",
		"BodyNode<[ForNode<SetOpNode<=, i, ConstNode<0, 0, 0>>, BinOpNode<lt, VarNode<i>, VarNode<n>>, UniOpNode<++, VarNode<i>>, BodyNode<[SetOpNode<=, a, ConstNode<0, 0, 0>>]>>]>")
	matchAst(t,
		"a = 1; a++ ",
		"BodyNode<[SetOpNode<=, a, ConstNode<0, 1, 0>> UniOpNode<++, VarNode<a>>]>")

	matchAst(t,
		"if (a == 0) { while (a == 0) { a++; } }",
		"BodyNode<[IfNode<BinOpNode<==, VarNode<a>, ConstNode<0, 0, 0>>, BodyNode<[WhileNode<BinOpNode<==, VarNode<a>, ConstNode<0, 0, 0>>, BodyNode<[UniOpNode<++, VarNode<a>>]>>]>, nil>]>")

	matchAst(t,
		"if (a) { a = 0; } else if (b) { b = 0; } else { c = 0; }",
		"BodyNode<[IfNode<VarNode<a>, BodyNode<[SetOpNode<=, a, ConstNode<0, 0, 0>>]>, IfNode<VarNode<b>, BodyNode<[SetOpNode<=, b, ConstNode<0, 0, 0>>]>, BodyNode<[SetOpNode<=, c, ConstNode<0, 0, 0>>]>>>]>")
}

func TestParseFnDef(t *testing.T) {
	matchAst(t,
		"func afn(a, b, c) { a = 0; }",
		"BodyNode<[FnDefNode<afn, [a b c], BodyNode<[SetOpNode<=, a, ConstNode<0, 0, 0>>]>>]>")
}

func TestParseOk3(t *testing.T) {
	matchAst(t,
		"2**(1/2)",
		"BodyNode<[BinOpNode<**, ConstNode<0, 2, 0>, BinOpNode</, ConstNode<0, 1, 0>, ConstNode<0, 2, 0>>>]>")
}

func TestParseError(t *testing.T) {
	pgm := "2 + * 3"
	_, err := parse(lex(strings.NewReader(pgm)))
	if (err == nil) || (err.Error() != "Syntax error: unexpected token '*' in line 1  (while parsing basic expression)") {
		t.Fatalf("Wrong or no error returned: %v\n", err)
	}
}

func TestParseMulDiv(t *testing.T) {
	matchAst(t,
		"11/25 * 2",
		"BodyNode<[BinOpNode<*, BinOpNode</, ConstNode<0, 11, 0>, ConstNode<0, 25, 0>>, ConstNode<0, 2, 0>>]>")
	matchAst(t,
		"2 * 11/25",
		"BodyNode<[BinOpNode<*, ConstNode<0, 2, 0>, BinOpNode</, ConstNode<0, 11, 0>, ConstNode<0, 25, 0>>>]>")
}

func TestAtSyntax(t *testing.T) {
	matchAst(t,
		"@",
		"BodyNode<[DpyNode<VarNode<_>>]>")
	matchAst(t,
		"@ p",
		"BodyNode<[DpyNode<VarNode<p>>]>")
	matchAst(t,
		"@:p",
		"BodyNode<[DpyNode<toggleProg>]>")
}
