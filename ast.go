package main

import (
	"fmt"
	"math/big"
	"time"
)

type BuiltinFunc func(argv []*value, lineno int) *value

type BuiltinFn struct {
	nargs int
	fn    BuiltinFunc
}

type AstNode interface {
	String() string
	Exec(callStack []CallFrame) *value
	Line() int
}

type BodyNode struct {
	statements []AstNode
	lineno     int
}

func NewBodyNode(statements []AstNode, lineno int) *BodyNode {
	return &BodyNode{statements, lineno}
}

func (n *BodyNode) String() string {
	return fmt.Sprintf("BodyNode<%s>", n.statements)
}

func (n *BodyNode) Line() int {
	return n.lineno
}

type UniOpNode struct {
	name   string
	fn     func(*value, int) *value
	child  AstNode
	lineno int
}

func NewUniOpNode(tok token, child AstNode) *UniOpNode {
	fn := tok.ttype.UniFn
	return &UniOpNode{tok.ttype.Name, fn, child, tok.lineno}
}

func (n *UniOpNode) String() string {
	return fmt.Sprintf("UniOpNode<%s, %s>", n.name, n.child.String())
}

func (n *UniOpNode) Line() int {
	return n.lineno
}

type VarNode struct {
	name   string
	lineno int
}

func NewVarNode(name string, lineno int) *VarNode {
	return &VarNode{name, lineno}
}

func (n *VarNode) String() string {
	return fmt.Sprintf("VarNode<%s>", n.name)
}

func (n *VarNode) Line() int {
	return n.lineno
}

type FnCallNode struct {
	name   string
	args   []AstNode
	lineno int
}

func NewFnCallNode(name string, args []AstNode, lineno int) *FnCallNode {
	return &FnCallNode{name, args, lineno}
}

func (n *FnCallNode) String() string {
	return fmt.Sprintf("FnCallNode<%s, %s>", n.name, n.args)
}

func (n *FnCallNode) Line() int {
	return n.lineno
}

type ConstNode struct {
	v      value
	lineno int
}

func NewConstNode(kind valueKind, flavor valueFlavor, ival int64, dval float64, lineno int) *ConstNode {
	r := &ConstNode{}
	r.v.kind = kind
	r.v.flavor = flavor
	r.v.ival = *big.NewInt(ival)
	r.v.dval = dval
	r.lineno = lineno
	return r
}

func NewDateNode(t time.Time, lineno int) *ConstNode {
	r := &ConstNode{}
	r.v.kind = DTVAL
	r.v.dtval = &t
	r.lineno = lineno
	return r
}

func (n *ConstNode) String() string {
	return fmt.Sprintf("ConstNode<%d, %s, %g>", n.v.kind, n.v.ival.String(), n.v.dval)
}

func (n *ConstNode) Line() int {
	return n.lineno
}

type BinOpFunc func(a1, a2 *value, kind valueKind, lineno int) *value

type BinOpNode struct {
	name     string
	fn       BinOpFunc
	op1      AstNode
	op2      AstNode
	lineno   int
	priority int
}

func NewBinOpNode(tok token, op1, op2 AstNode) *BinOpNode {
	fn := tok.ttype.BinFn
	return &BinOpNode{tok.ttype.Name, fn, op1, op2, tok.lineno, tok.ttype.Priority}
}

func (n *BinOpNode) String() string {
	return fmt.Sprintf("BinOpNode<%s, %s, %s>", n.name, n.op1.String(), n.op2.String())
}

func (n *BinOpNode) Line() int {
	return n.lineno
}

type SetOpNode struct {
	name    string
	fnOp    BinOpFunc
	varName string
	op1     AstNode
	lineno  int
}

func NewSetOpNode(tok token, varName string, op1 AstNode) *SetOpNode {
	return &SetOpNode{tok.val, tok.ttype.BinFn, varName, op1, tok.lineno}
}

func (n *SetOpNode) String() string {
	return fmt.Sprintf("SetOpNode<%s, %s, %s>", n.name, n.varName, n.op1)
}

func (n *SetOpNode) Line() int {
	return n.lineno
}

type WhileNode struct {
	guard  AstNode
	body   AstNode
	lineno int
}

func NewWhileNode(guard, body AstNode, lineno int) *WhileNode {
	return &WhileNode{guard, body, lineno}
}

func (n *WhileNode) String() string {
	return fmt.Sprintf("WhileNode<%s, %s>", n.guard, n.body)
}

func (n *WhileNode) Line() int {
	return n.lineno
}

type ForNode struct {
	initExpr AstNode
	guard    AstNode
	incrExpr AstNode
	body     AstNode
	lineno   int
}

func NewForNode(initExpr, guard, incrExpr, body AstNode, lineno int) AstNode {
	return &ForNode{initExpr, guard, incrExpr, body, lineno}
}

func (n *ForNode) String() string {
	return fmt.Sprintf("ForNode<%s, %s, %s, %s>", n.initExpr, n.guard, n.incrExpr, n.body)
}

func (n *ForNode) Line() int {
	return n.lineno
}

type IfNode struct {
	guard    AstNode
	ifBody   AstNode
	elseBody AstNode
	lineno   int
}

func NewIfNode(guard, ifBody, elseBody AstNode, lineno int) AstNode {
	return &IfNode{guard, ifBody, elseBody, lineno}
}

func (n *IfNode) String() string {
	elseStr := "nil"
	if n.elseBody != nil {
		elseStr = n.elseBody.String()
	}
	return fmt.Sprintf("IfNode<%s, %s, %s>", n.guard, n.ifBody, elseStr)
}

func (n *IfNode) Line() int {
	return n.lineno
}

type FnDefNode struct {
	name   string
	args   []string
	body   AstNode
	lineno int
}

func NewFnDefNode(name string, args []string, body AstNode, lineno int) *FnDefNode {
	return &FnDefNode{name, args, body, lineno}
}

func (n *FnDefNode) String() string {
	return fmt.Sprintf("FnDefNode<%s, %s, %s>", n.name, n.args, n.body)
}

func (n *FnDefNode) Line() int {
	return n.lineno
}

type NilNode struct {
}

func NewNilNode() NilNode {
	return NilNode{}
}

func (n *NilNode) String() string {
	return "!!!NILNODE!!!"
}

func (n *NilNode) Line() int {
	return -1
}

type DpyNode struct {
	expr       AstNode
	toggleProg bool
	lineno     int
}

func (n *DpyNode) String() string {
	if n.toggleProg {
		return "DpyNode<toggleProg>"
	} else {
		return fmt.Sprintf("DpyNode<%s>", n.expr)
	}
}

func (n *DpyNode) Line() int {
	return n.lineno
}

type ExitNode struct {
	lineno int
}

func (n *ExitNode) String() string {
	return fmt.Sprintf("ExitNode<>")
}

func (n *ExitNode) Line() int {
	return n.lineno
}
