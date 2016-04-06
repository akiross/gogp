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
