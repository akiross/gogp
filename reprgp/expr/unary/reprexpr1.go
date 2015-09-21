package unary

import "ale-re.net/phd/gogp"

type NumericIn int
type NumericOut float64

type Terminal func(x NumericIn) NumericOut // For expressions in 1 variable

type Functional1 func(args ...gogp.Primitive) gogp.Primitive // Unary operations
type Functional2 func(args ...gogp.Primitive) gogp.Primitive // Binary operations

// The following are to satisfy the interface
func (self Terminal) IsFunctional() bool                     { return false }
func (self Terminal) Arity() int                             { return -1 }
func (self Terminal) Run(p ...gogp.Primitive) gogp.Primitive { return self }

func (self Functional1) IsFunctional() bool                     { return true }
func (self Functional1) Arity() int                             { return 1 }
func (self Functional1) Run(p ...gogp.Primitive) gogp.Primitive { return self(p[0]) }

func (self Functional2) IsFunctional() bool                     { return true }
func (self Functional2) Arity() int                             { return 2 }
func (self Functional2) Run(p ...gogp.Primitive) gogp.Primitive { return self(p[0], p[1]) }

func Identity(x NumericIn) NumericOut {
	return x
}

func Constant1(c NumericIn) Terminal {
	return func(_ NumericIn) NumericOut {
		return c
	}
}

func Sum(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		return args[0].(Terminal)(x) + args[1].(Terminal)(x)
	})
}

func Sub(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		return args[0].(Terminal)(x) - args[1].(Terminal)(x)
	})
}

func Mul(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		return args[0].(Terminal)(x) * args[1].(Terminal)(x)
	})
}

func ProtectedDiv(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		n, d := args[0].(Terminal)(x), args[1].(Terminal)(x)
		if d == NumericCout(0) {
			return NumericOut(1)
		} else {
			return n / d
		}
	})
}

func Square(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		v := args[0].(Terminal)(x)
		return v * v
	})
}

func Abs(args ...gogp.Primitive) gogp.Primitive {
	return Terminal(func(x NumericIn) NumericOut {
		v := args[0].(Terminal)(x)
		if v < 0 {
			return -v
		} else {
			return v
		}
	})
}
