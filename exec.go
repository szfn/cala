package main

import (
	"fmt"
	"io"
)

type CallFrame struct {
	vars map[string]*value
}

type ExecError struct {
	msg string
	stackTrace []string
}

func NewExecError(msg string) *ExecError {
	return &ExecError{msg, getStackTrace()}
}

func (ee *ExecError) printStackTrace(out io.Writer) {
	for _, frame := range ee.stackTrace {
		fmt.Fprintf(out, "at %s\n", frame)
	}
}

func (ee *ExecError) Error() string {
	return ee.msg
}

// Executes the specified (parsed) program
func execWithCallStack(program AstNode, callStack []CallFrame) (v *value, err error) {
	defer func() {
		if p := recover(); p != nil {
			if err2, ok := p.(error); ok {
				err = NewExecError(err2.Error())
			} else {
				panic(p)
			}
		}
	}()

	v = program.Exec(callStack)
	return
}

func exec(program AstNode) (v *value, err error) {
	return execWithCallStack(program, NewCallStack())
}

func NewCallStack() []CallFrame {
	return []CallFrame{
		{
			vars: map[string]*value {
				"abs": btnAbs,
				"acos": btnAcos,
				"asin": btnAsin,
				"atan": btnAtan,
				"cos": btnCos,
				"cosh": btnCosh,
				"floor": btnFloor,
				"ceil": btnCeil,
				"ln": btnLn,
				"log10": btnLog10,
				"log2": btnLog2,
				"sin": btnSin,
				"sinh": btnSinh,
				"sqrt": btnSqrt,
				"tan": btnTan,
				"tanh": btnTanh,
				"dpy": btnDpy,
			},
		},
	}
}

// looks up the value of a variable, note that we implement *static* scoping:
// A variable must have been defined in the local scope of a function or in the local scope
// there is no lexical or dynamic scoping
// If alsoDefine is specified and the variable is not found a new one with that name is created
// If alsoDefine is false and the variable is not found lookup panics
func lookup(stack []CallFrame, name string, alsoDefine bool, lineno int) *value {
	frame := stack[len(stack)-1]
	vv, ok := frame.vars[name]
	if !ok {
		// lookup global call frame instead
		vv, ok = stack[0].vars[name]
		if !ok {
			if alsoDefine {
				vv = &value{}
				frame.vars[name] = vv
				return vv
			} else {
				panic(fmt.Errorf("Unknown variable %s at line %d", name, lineno))
			}
		} else {
			return vv
		}
	}
	return vv
}

func (n *BodyNode) Exec(stack []CallFrame) (vv *value) {
	for _, stmt := range n.statements {
		vv = stmt.Exec(stack)
	}
	return
}

func (n *UniOpNode) Exec(stack []CallFrame) *value {
	a := n.child.Exec(stack)
	return n.fn(a, n.lineno)
}

func (n *VarNode) Exec(stack []CallFrame) *value {
	return lookup(stack, n.name, false, n.lineno)
}

func (n *FnCallNode) Exec(stack []CallFrame) *value {
	// retrieves function definition
	vv := lookup(stack, n.name, false, n.lineno)

	// evaluates arguments
	argv := make([]*value, len(n.args))
	for i, arg := range n.args {
		argv[i] = arg.Exec(stack)
	}

	switch vv.kind {
	case PVAL:
		return functionCall(n, vv.nval, argv, stack)

	case BVAL:
		if vv.bval == nil {
			panic(fmt.Errorf("Can not call '%s' (internal error) at line %d", n.name, n.lineno))
		}
		if vv.bval.nargs != len(argv) {
			panic(fmt.Errorf("Can not call '%s' at line %d: wrong number of arguments", n.name, n.lineno))
		}
		return vv.bval.fn(argv, n.lineno)
	}
	panic(fmt.Errorf("Can not call '%s' at line %d: not a function", n.name, n.lineno))
}

// Calls a user defined function: n is the call node, fn is the function definition node, argv are values to pass as arguments
func functionCall(n *FnCallNode, fn *FnDefNode, argv []*value, stack []CallFrame) *value {
	if fn == nil {
		panic(fmt.Errorf("Can not call '%s' (internal error) at line %d", n.name, n.lineno))
	}
	if len(fn.args) != len(argv) {
		panic(fmt.Errorf("Can not call '%s' at line %d: wrong number of arguments (given %d expected %d)", n.name, n.lineno, len(argv), len(fn.args)))
	}

	stack = append(stack, CallFrame{
		vars: map[string]*value{},
	})

	newFrame := &stack[len(stack)-1]

	for i, arg := range fn.args {
		newFrame.vars[arg] = argv[i]
	}

	retv := fn.body.Exec(stack)

	stack = stack[:len(stack)-1]

	return retv
}


func (n *ConstNode) Exec(stack []CallFrame) *value {
	if (n.v.kind == BVAL) || (n.v.kind == PVAL) {
		panic(fmt.Errorf("Internal error, a literal function appeared at line %d", n.lineno))
	}
	vv := n.v
	return &vv
}

func (n *BinOpNode) Exec(stack []CallFrame) *value {
	a1 := n.op1.Exec(stack)
	a2 := n.op2.Exec(stack)
	return n.fn(a1, a2, n.lineno)
}

func (n *SetOpNode) Exec(stack []CallFrame) *value {
	alsoDefine := (n.name == "=")
	a1 := lookup(stack, n.varName, alsoDefine, n.lineno)
	a2 := n.op1.Exec(stack)
	vv := a2
	if n.fnOp != nil {
		vv = n.fnOp(a1, a2, n.lineno)
	}
	*a1 = *vv
	vvv := *a1
	return &vvv
}

func (n *WhileNode) Exec(stack []CallFrame) (vv *value){
	vv = &value{ IVAL, 0, 0.0, nil, nil }
	for {
		gv := n.guard.Exec(stack)
		if !gv.Bool(n.guard.Line()) {
			break
		}
		vv = n.body.Exec(stack)
	}
	return
}

func (n *ForNode) Exec(stack []CallFrame) (vv *value) {
	vv = &value{ IVAL, 0, 0.0, nil, nil }

	n.initExpr.Exec(stack)

	for {
		gv := n.guard.Exec(stack)
		if !gv.Bool(n.guard.Line()) {
			break
		}
		vv = n.body.Exec(stack)
		n.incrExpr.Exec(stack)
	}

	return
}

func (n *IfNode) Exec(stack []CallFrame) *value {
	ev := n.guard.Exec(stack)

	if ev.Bool(n.guard.Line()) {
		return n.ifBody.Exec(stack)
	} else {
		if n.elseBody != nil {
			return n.elseBody.Exec(stack)
		}
	}
	return &value{ IVAL, 0, 0, nil, nil }
}

func (n *FnDefNode) Exec(stack []CallFrame) *value {
	frame := stack[len(stack)-1]
	vv := &value{ PVAL, 0, 0, n, nil }
	frame.vars[n.name] = vv
	return vv
}

func asInt(x bool) int64 {
	if x {
		return 1
	}
	return 0
}

func (vv *value) Bool(lineno int) bool {
	switch vv.kind {
	case IVAL:
		return vv.ival != 0
	case DVAL:
		panic(fmt.Errorf("Real value can not be used as boolean at line %d", lineno))
	default:
		panic(fmt.Errorf("Function value can not be used as boolean at line %d\n", lineno))
	}
	panic("Unreachable")
}

func (vv *value) Int(lineno int) int64 {
	if vv.kind != IVAL {
		panic(fmt.Errorf("Can not use non-integer value as integer at line %d\n", lineno))
	}
	return vv.ival
}

func (vv *value) Real(lineno int) float64 {
	switch vv.kind {
	case IVAL:
		return float64(vv.ival)
	case DVAL:
		return vv.dval
	}
	panic(fmt.Errorf("Can not use non-number value as real at line %d", lineno))
}

func (vv *value) String() string {
	switch vv.kind {
	case IVAL:
		return fmt.Sprintf("%d", vv.ival)
	case DVAL:
		return fmt.Sprintf("%g", vv.dval)
	}
	return fmt.Sprintf("@")
}
