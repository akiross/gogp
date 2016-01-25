package rr

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

// This is what will render something on screen
type RenderFunc func(x1, x2, y1, y2 float64, img *imgut.Image)

type Primitive struct {
	name       string
	functional bool
	arity      int
	Render     RenderFunc
}

// Returns true if is functional, false if terminal
func (p *Primitive) IsFunctional() bool {
	return p.functional
}

func (p *Primitive) Arity() int {
	return p.arity
}

// Functionals returns terminals, terminals do nothing
func (p *Primitive) Run(args ...gp.Primitive) gp.Primitive {
	if p.name == "VSplit" {
		return &Primitive{"VSplit", false, -1, func(x1, y1, x2, y2 float64, img *imgut.Image) {
			xh := (x1 + x2) * 0.5
			left := args[0].(*Primitive)
			right := args[1].(*Primitive)
			left.Render(x1, y1, xh, y2, img)
			right.Render(xh, y1, x2, y2, img)
		}}
	} else if p.name == "HSplit" {
		return &Primitive{"HSplit", false, -1, func(x1, y1, x2, y2 float64, img *imgut.Image) {
			yh := (y1 + y2) * 0.5
			args[0].(*Primitive).Render(x1, y1, x2, yh, img)
			args[1].(*Primitive).Render(x1, yh, x2, y2, img)
		}}
	} else {
		return p
	}
}

func (p *Primitive) Name() string {
	return p.name
}

func MakeVSplit() *Primitive {
	return &Primitive{"VSplit", true, 2, nil}
}

func MakeHSplit() *Primitive {
	return &Primitive{"HSplit", true, 2, nil}
}

func MakeTerminal(name string, rf RenderFunc) *Primitive {
	return &Primitive{name, false, -1, rf}
}

func Filler(col ...float64) RenderFunc {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, col...)
	}
}

// Builds a Terminal that fill the entire area with a shaded line
func LinShade(startCol, endCol, sx, sy, ex, ey float64) RenderFunc {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.LinearShade(x1, y1, x2, y2, sx, sy, ex, ey, startCol, endCol)
	}
}
