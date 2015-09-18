package gogp

import (
	"fmt"
	"testing"
)

/// Some testing primitives

type Terminal1 func(x int) int    // For expressions in 1 variable
type Terminal2 func(x, y int) int // For expressions in 2 variables

type Functional1 func(args ...Primitive) Primitive // Unary operations
type Functional2 func(args ...Primitive) Primitive // Binary operations

// The following are to satisfy the interface
func (self Terminal1) IsFunctional() bool           { return false }
func (self Terminal1) Arity() int                   { return -1 }
func (self Terminal1) Run(p ...Primitive) Primitive { return self }

func (self Terminal2) IsFunctional() bool           { return false }
func (self Terminal2) Arity() int                   { return -1 }
func (self Terminal2) Run(p ...Primitive) Primitive { return self }

func (self Functional1) IsFunctional() bool           { return true }
func (self Functional1) Arity() int                   { return 1 }
func (self Functional1) Run(p ...Primitive) Primitive { return self(p[0]) }

func (self Functional2) IsFunctional() bool           { return true }
func (self Functional2) Arity() int                   { return 2 }
func (self Functional2) Run(p ...Primitive) Primitive { return self(p[0], p[1]) }

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

func Sum(args ...Primitive) Primitive {
	fmt.Println("Running Run for the Sum, got terminals", args, "I am", Sum)
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) + args[1].(Terminal1)(x)
	})
}

func Sub(args ...Primitive) Primitive {
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) - args[1].(Terminal1)(x)
	})
}

func Abs(args ...Primitive) Primitive {
	return Terminal1(func(x int) int {
		v := args[0].(Terminal1)(x)
		if v < 0 {
			return -v
		} else {
			return v
		}
	})
}

var functionals []Primitive = []Primitive{Functional1(Sum), Functional1(Sub), Functional1(Abs)}
var terminals []Primitive = []Primitive{Terminal1(Constant1(0)), Terminal1(Constant1(1)), Terminal1(Constant1(-1))}

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
		t1 := MakeTreeHalfAndHalf(maxDepth, functionals, terminals)
		t2 := MakeTreeHalfAndHalf(maxDepth, functionals, terminals)
		t3 := MakeTreeHalfAndHalf(maxDepth+1, functionals, terminals)
		t4 := MakeTreeHalfAndHalf(maxDepth+1, functionals, terminals)
		t5 := MakeTreeHalfAndHalf(maxDepth+2, functionals, terminals)
		t6 := MakeTreeHalfAndHalf(maxDepth+2, functionals, terminals)

		// Repeat M times each crossover
		for j := 0; j < M; j++ {
			d1, d2 := t1.Depth(), t2.Depth()
			xo0(t1, t2)
			if t1.Depth() > maxDepth || t2.Depth() > maxDepth {
				t.Errorf("xo0 got d1: %v d2: %v after crossover d1': %v d2': %v", d1, d2, t1.Depth(), t2.Depth())
			}

			d3, d4 := t3.Depth(), t4.Depth()
			xo1(t3, t4)
			if t3.Depth() > maxDepth+1 || t4.Depth() > maxDepth+1 {
				t.Errorf("xo1 got d3: %v d4: %v after crossover d3': %v d4': %v", d3, d4, t3.Depth(), t4.Depth())
			}

			d5, d6 := t5.Depth(), t6.Depth()
			xo2(t5, t6)
			if t5.Depth() > maxDepth+2 || t6.Depth() > maxDepth+2 {
				t.Errorf("xo3 got d5: %v d6: %v after crossover d5': %v d6': %v", d5, d6, t5.Depth(), t6.Depth())
			}

		}
	}
}

func mt(p Primitive, children ...*Node) *Node {
	return &Node{p, children}
}

func TestEvaluation(t *testing.T) {
	// Explicitly make a tree: 1 + |0 - x|
	zero, one, id := Terminal1(Constant1(0)), Terminal1(Constant1(1)), Terminal1(Identity1)
	sum, abs, sub := Functional1(Sum), Functional1(Abs), Functional1(Sub)
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

	if v := eSum(-1); v != 2 {
		t.Error("expression 'Sum(1, 1)' should have value 2 but had", v)
	}
}
