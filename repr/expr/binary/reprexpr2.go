package binary

import (
	"github.com/akiross/gogp/gp"
	"math"
)

type NumericIn int
type NumericOut float64

type Terminal func(x, y NumericIn) NumericOut

type Functional1 func(args ...gp.Primitive) gp.Primitive // Unary operations
type Functional2 func(args ...gp.Primitive) gp.Primitive // Binary operations
type Functional3 func(args ...gp.Primitive) gp.Primitive // Ternary operations

// The following are to satisfy the interface
func (self Terminal) IsFunctional() bool                 { return false }
func (self Terminal) Arity() int                         { return -1 }
func (self Terminal) Run(p ...gp.Primitive) gp.Primitive { return self }
func (self Terminal) Name() string                       { return "Terminal" }

func (self Functional1) IsFunctional() bool                 { return true }
func (self Functional1) Arity() int                         { return 1 }
func (self Functional1) Run(p ...gp.Primitive) gp.Primitive { return self(p[0]) }
func (self Functional1) Name() string                       { return "Terminal" }

func (self Functional2) IsFunctional() bool                 { return true }
func (self Functional2) Arity() int                         { return 2 }
func (self Functional2) Run(p ...gp.Primitive) gp.Primitive { return self(p[0], p[1]) }
func (self Functional2) Name() string                       { return "Terminal" }

func (self Functional3) IsFunctional() bool                 { return true }
func (self Functional3) Arity() int                         { return 3 }
func (self Functional3) Run(p ...gp.Primitive) gp.Primitive { return self(p[0], p[1], p[2]) }
func (self Functional3) Name() string                       { return "Terminal" }

func IdentityX(x, _ NumericIn) NumericOut {
	return NumericOut(x)
}

func IdentityY(_, y NumericIn) NumericOut {
	return NumericOut(y)
}

func Constant(c NumericOut) Terminal {
	return func(_, _ NumericIn) NumericOut {
		return c
	}
}

func Sum(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		return args[0].(Terminal)(x, y) + args[1].(Terminal)(x, y)
	})
}

func Sub(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		return args[0].(Terminal)(x, y) - args[1].(Terminal)(x, y)
	})
}

func Mul(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		return args[0].(Terminal)(x, y) * args[1].(Terminal)(x, y)
	})
}

func ProtectedDiv(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		n, d := args[0].(Terminal)(x, y), args[1].(Terminal)(x, y)
		if d == NumericOut(0) {
			return NumericOut(1)
		} else {
			return n / d
		}
	})
}

func Square(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v := args[0].(Terminal)(x, y)
		v = v * v
		if v < -1e6 {
			return -1e6
		} else if v > 1e6 {
			return 1e6
		} else {
			return v
		}
	})
}

func Min(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v, w := args[0].(Terminal)(x, y), args[1].(Terminal)(x, y)
		if v < w {
			return v
		} else {
			return w
		}
	})
}

func Max(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v, w := args[0].(Terminal)(x, y), args[1].(Terminal)(x, y)
		if v > w {
			return v
		} else {
			return w
		}
	})
}

func Pow(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		base, exp := args[0].(Terminal)(x, y), args[1].(Terminal)(x, y)
		return NumericOut(math.Pow(float64(base), float64(exp)))
	})
}

func Abs(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		if v := args[0].(Terminal)(x, y); v < 0 {
			return -v
		} else {
			return v
		}
	})
}

func Neg(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		return -args[0].(Terminal)(x, y)
	})
}

func Sign(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		if args[0].(Terminal)(x, y) < 0 {
			return -1
		} else {
			return 1
		}
	})
}

func Sqrt(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		return NumericOut(math.Sqrt(float64(args[0].(Terminal)(x, y))))
	})
}

func Choice(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		if args[0].(Terminal)(x, y) > 0 {
			return args[1].(Terminal)(x, y)
		} else {
			return args[2].(Terminal)(x, y)
		}
	})
}
