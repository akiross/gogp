package binary

import (
	"fmt"
	"github.com/akiross/gogp/gp"
	//	"math"
)

type NumericIn float64
type NumericOut float64

type Primitive struct {
	name       string
	functional bool
	arity      int
	Eval       func(x, y NumericIn) NumericOut
	ephemeral  func() *Primitive
	compose    func(args ...gp.Primitive) *Primitive
}

func (p *Primitive) Name() string {
	return p.name
}

// Returns true if is functional, false if terminal
func (p *Primitive) IsFunctional() bool {
	return p.functional
}

func (p *Primitive) IsEphemeral() bool {
	return p.ephemeral != nil
}

func (p *Primitive) Arity() int {
	return p.arity
}

func (p *Primitive) Run(args ...gp.Primitive) gp.Primitive {
	if p.Arity() > 0 {
		return p.compose(args...)
	} else if p.IsEphemeral() {
		return p.ephemeral()
	} else {
		return p // Terminal
	}
}

/*
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

*/

func MakeIdentityX() *Primitive {
	return &Primitive{"IdX", false, -1, func(x, _ NumericIn) NumericOut {
		return NumericOut(x)
	}, nil, nil}
}

func MakeIdentityY() *Primitive {
	return &Primitive{"IdY", false, -1, func(_, y NumericIn) NumericOut {
		return NumericOut(y)
	}, nil, nil}
}

func MakeConstant(c NumericOut) *Primitive {
	return &Primitive{fmt.Sprintf("C_%v", c), false, -1, func(_, _ NumericIn) NumericOut {
		return c
	}, nil, nil}
}

func MakeEphimeral(name string, gen func() *Primitive) *Primitive {
	return &Primitive{name, false, -1, nil, gen, nil}
}

func MakeUnary(name string, oper func(a NumericOut) NumericOut) *Primitive {
	return &Primitive{name, true, 1, nil, nil, func(a ...gp.Primitive) *Primitive {
		a1 := a[0].(*Primitive)
		return &Primitive{name, false, -1, func(x, y NumericIn) NumericOut {
			return oper(a1.Eval(x, y))
		}, nil, nil}
	}}
}

func MakeBinary(name string, oper func(a, b NumericOut) NumericOut) *Primitive {
	return &Primitive{name, true, 2, nil, nil, func(a ...gp.Primitive) *Primitive {
		a1, a2 := a[0].(*Primitive), a[1].(*Primitive)
		return &Primitive{name, false, -1, func(x, y NumericIn) NumericOut {
			return oper(a1.Eval(x, y), a2.Eval(x, y))
		}, nil, nil}
	}}
}

func MakeTernary(name string, oper func(a, b, c NumericOut) NumericOut) *Primitive {
	return &Primitive{name, true, 3, nil, nil, func(a ...gp.Primitive) *Primitive {
		a1, a2, a3 := a[0].(*Primitive), a[1].(*Primitive), a[2].(*Primitive)
		return &Primitive{name, false, -1, func(x, y NumericIn) NumericOut {
			return oper(a1.Eval(x, y), a2.Eval(x, y), a3.Eval(x, y))
		}, nil, nil}
	}}
}

/*

	return &Primitive{"VSplit", false, -1, func(x1, y1, x2, y2 float64, img *imgut.Image) {
		xh := (x1 + x2) * 0.5
		left := args[0].(*Primitive)
		right := args[1].(*Primitive)
		left.Render(x1, y1, xh, y2, img)
		right.Render(xh, y1, x2, y2, img)
	}, nil}
*/

/*

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

func Sin(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v := args[0].(Terminal)(x, y)
		return math.Sin(v)
	})
}

func Cos(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v := args[0].(Terminal)(x, y)
		return math.Cos(v)
	})
}

func Tanh(args ...gp.Primitive) gp.Primitive {
	return Terminal(func(x, y NumericIn) NumericOut {
		v := args[0].(Terminal)(x, y)
		return math.Tanh(v)
	})
}
*/
