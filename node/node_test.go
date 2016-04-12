package node

import (
	"github.com/akiross/gogp/gp"
	"testing"
)

type Terminal1 func(x int) int    // For expressions in 1 variable
type Terminal2 func(x, y int) int // For expressions in 2 variables

type Functional1 func(args ...gp.Primitive) gp.Primitive // Unary operations
type Functional2 func(args ...gp.Primitive) gp.Primitive // Binary operations

// The following are to satisfy the interface
func (self Terminal1) IsFunctional() bool                 { return false }
func (self Terminal1) Arity() int                         { return -1 }
func (self Terminal1) Run(p ...gp.Primitive) gp.Primitive { return self }
func (self Terminal1) IsEphemeral() bool                  { return false }
func (self Terminal1) Name() string                       { return gp.FuncName(self) }

func (self Terminal2) IsFunctional() bool                 { return false }
func (self Terminal2) Arity() int                         { return -1 }
func (self Terminal2) Run(p ...gp.Primitive) gp.Primitive { return self }
func (self Terminal2) IsEphemeral() bool                  { return false }
func (self Terminal2) Name() string                       { return gp.FuncName(self) }

func (self Functional1) IsFunctional() bool                 { return true }
func (self Functional1) Arity() int                         { return 1 }
func (self Functional1) Run(p ...gp.Primitive) gp.Primitive { return self(p[0]) }
func (self Functional1) IsEphemeral() bool                  { return false }
func (self Functional1) Name() string                       { return gp.FuncName(self) }

func (self Functional2) IsFunctional() bool                 { return true }
func (self Functional2) Arity() int                         { return 2 }
func (self Functional2) Run(p ...gp.Primitive) gp.Primitive { return self(p[0], p[1]) }
func (self Functional2) IsEphemeral() bool                  { return false }
func (self Functional2) Name() string                       { return gp.FuncName(self) }

func Identity1(x int) int {
	return x
}

func Constant1(c int) Terminal1 {
	return func(_ int) int {
		return c
	}
}

func Constant2(c int) Terminal2 {
	return func(_, _ int) int {
		return c
	}
}

func Sum(args ...gp.Primitive) gp.Primitive {
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) + args[1].(Terminal1)(x)
	})
}

func Sub(args ...gp.Primitive) gp.Primitive {
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) - args[1].(Terminal1)(x)
	})
}

func Abs(args ...gp.Primitive) gp.Primitive {
	return Terminal1(func(x int) int {
		v := args[0].(Terminal1)(x)
		if v < 0 {
			return -v
		} else {
			return v
		}
	})
}

var functionals []gp.Primitive = []gp.Primitive{Functional1(Sum), Functional1(Sub), Functional1(Abs)}
var terminals []gp.Primitive = []gp.Primitive{Terminal1(Constant1(0)), Terminal1(Constant1(1)), Terminal1(Constant1(-1))}

func TestCrossover(t *testing.T) {
	// Testing crossover, ensuring limits are respected

	// Build two random trees
	maxDepth := 8

	xo0 := MakeTree1pCrossover(maxDepth)
	xo1 := MakeTree1pCrossover(maxDepth + 1)
	xo2 := MakeTree1pCrossover(maxDepth + 2)

	// Repeat N times with random trees
	N, M := 100, 100
	for i := 0; i < N; i++ {
		t1 := MakeTreeHalfAndHalf(0, maxDepth, functionals, terminals)
		t2 := MakeTreeHalfAndHalf(0, maxDepth, functionals, terminals)
		t3 := MakeTreeHalfAndHalf(0, maxDepth+1, functionals, terminals)
		t4 := MakeTreeHalfAndHalf(0, maxDepth+1, functionals, terminals)
		t5 := MakeTreeHalfAndHalf(0, maxDepth+2, functionals, terminals)
		t6 := MakeTreeHalfAndHalf(0, maxDepth+2, functionals, terminals)

		// Repeat M times each crossover
		for j := 0; j < M; j++ {
			d1, d2 := Depth(t1), Depth(t2)
			xo0(t1, t2)
			if Depth(t1) > maxDepth || Depth(t2) > maxDepth {
				t.Errorf("xo0 got d1: %v d2: %v after crossover d1': %v d2': %v", d1, d2, Depth(t1), Depth(t2))
			}

			d3, d4 := Depth(t3), Depth(t4)
			xo1(t3, t4)
			if Depth(t3) > maxDepth+1 || Depth(t4) > maxDepth+1 {
				t.Errorf("xo1 got d3: %v d4: %v after crossover d3': %v d4': %v", d3, d4, Depth(t3), Depth(t4))
			}

			d5, d6 := Depth(t5), Depth(t6)
			xo2(t5, t6)
			if Depth(t5) > maxDepth+2 || Depth(t6) > maxDepth+2 {
				t.Errorf("xo3 got d5: %v d6: %v after crossover d5': %v d6': %v", d5, d6, Depth(t5), Depth(t6))
			}

		}
	}
}

func mt(p gp.Primitive, children ...*Node) *Node {
	return &Node{p, children}
}

func TestEvaluation(t *testing.T) {
	// Explicitly make trees to test
	zero, one, id := Terminal1(Constant1(0)), Terminal1(Constant1(1)), Terminal1(Identity1)
	sum, abs, sub := Functional2(Sum), Functional1(Abs), Functional2(Sub)
	_, _, _, _, _, _ = zero, one, id, sum, abs, sub
	var tZero, tOne, tId *Node = mt(zero), mt(one), mt(id) //mt(sum, mt(one), mt(abs, mt(sub, mt(zero), mt(id))))
	//	var tOne *Node = mt(one)
	//	var tId *Node = mt(id)

	// Test constants and identity
	eZero := CompileTree(tZero).(Terminal1)
	eOne := CompileTree(tOne).(Terminal1)
	eId := CompileTree(tId).(Terminal1)

	if v := eZero(-1); v != 0 {
		t.Error("expression '0' should have value 0 but had", v)
	}
	if v := eOne(-1); v != 1 {
		t.Error("expression '1' should have value 1 but had", v)
	}
	if v := eId(-1); v != -1 {
		t.Error("expression 'Id(-1)' should have value -1 but had", v)
	}
	if v := eId(42); v != 42 {
		t.Error("expression 'Id(42)' should have value 42 but had", v)
	}

	// Test simple expressions
	tSum := mt(sum, tOne, tOne)
	eSum := CompileTree(tSum).(Terminal1)

	tSub := mt(sub, tZero, tId)
	eSub := CompileTree(tSub).(Terminal1)

	tAbs := mt(abs, tId)
	eAbs := CompileTree(tAbs).(Terminal1)

	if v := eSum(-1); v != 2 {
		t.Error("expression 'Sum(1, 1)' should have value 2 but had", v)
	}
	if v := eSub(0); v != 0 {
		t.Error("expression 'Sub(0, 0)' should have value 0 but had", v)
	}
	if v := eSub(2); v != -2 {
		t.Error("expression 'Sub(0, -2)' should have value -2 but had", v)
	}
	if v := eAbs(42); v != 42 {
		t.Error("expression 'Abs(42)' should have value 42 but had", v)
	}
	if v := eAbs(-42); v != 42 {
		t.Error("expression 'Abs(-42)' should have value 42 but had", v)
	}

	// Compound expression, arity 1 and 2 mixed   a tree: 2 + |0 - x|
	tFun := mt(sum, tSum, mt(abs, tSub))
	eFun := CompileTree(tFun).(Terminal1)

	if v := eFun(1); v != 3 {
		t.Error("expression 'Sum(Sum(1, 1), Abs(Sub(0, 1)))' should have value 3 but had", v)
	}
	if v := eFun(-1); v != 3 {
		t.Error("expression 'Sum(Sum(1, 1), Abs(Sub(0, -1)))' should have value 3 but had", v)
	}
	if v := eFun(2); v != 4 {
		t.Error("expression 'Sum(Sum(1, 1), Abs(Sub(0, 2)))' should have value 4 but had", v)
	}
	if v := eFun(-4); v != 6 {
		t.Error("expression 'Sum(Sum(1, 1), Abs(Sub(0, -4)))' should have value 6 but had", v)
	}
	if v := eFun(0); v != 2 {
		t.Error("expression 'Sum(Sum(1, 1), Abs(Sub(0, 0)))' should have value 2 but had", v)
	}
}
