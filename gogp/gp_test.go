package gogp

import (
	"testing"
)

// Some testing primitives

// FIXME qui c'é un problema grosso: che non è possibile gestire arità diverse nello stesso set!! CAZZO!

type Terminal1 func(x int, acc *int)
type Terminal2 func(x, y int, acc *int)

// Functionals are internal nodes of a solution tree
type Functional1 func(args ...Terminal) Terminal1
type Functional2 func(args ...Terminal) Terminal2

func (self Terminal1) IsFunctional() bool           { return false }
func (self Terminal1) Arity() int                   { return -1 }
func (self Terminal1) Run(p ...Primitive) Primitive { return self }

func (self Terminal2) IsFunctional() bool           { return false }
func (self Terminal2) Arity() int                   { return -1 }
func (self Terminal2) Run(p ...Primitive) Primitive { return self }

func (self Functional1) IsFunctional() bool           { return true }
func (self Functional1) Arity() int                   { return 1 }
func (self Functional1) Run(p ...Primitive) Primitive { return self(p[0].(Terminal)) }

func (self Functional2) IsFunctional() bool           { return true }
func (self Functional2) Arity() int                   { return 2 }
func (self Functional2) Run(p ...Primitive) Primitive { return self(p[0].(Terminal), p[1].(Terminal)) }

func Constant1(c int) Terminal1 {
	return func(x int, acc *int) {
		*acc = c
	}
}

func Sum(args ...Terminal) Terminal {
	return func(x, y int, acc *int) {
		*acc = x + y
	}
}

func Sub(args ...Terminal) Terminal {
	return func(x, y int, acc *int) {
		*acc = x - y
	}
}

func Minus(args ...Terminal) Terminal {
	return func(x int, acc *int) {
		*acc = -x
	}
}

func TestCrossover(t *testing.T) {
	// Testing crossover, ensuring limits are respected

	functionals := []Primitive{Functional(Sum), Functional(Sub), Functional(Minus)}
	terminals := []Primitive{Terminal(Constant(0)), Terminal(Constant(1)), Terminal(Constant(-1))}

	// Build two random trees
	maxDepth := 7

	xo0 := MakeTree1pCrossover(maxDepth)
	xo1 := MakeTree1pCrossover(maxDepth + 1)
	xo2 := MakeTree1pCrossover(maxDepth + 2)

	// Repeat N times with random trees
	for i := 0; i < 10; i++ {
		//		t1 := MakeTreeHalfAndHalf(maxDepth, functionals, terminals)
		//		t2 := MakeTreeHalfAndHalf(maxDepth, functionals, terminals)

		// Repeat M times each crossover
		//		t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
	}
}
