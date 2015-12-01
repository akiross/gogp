package vhs

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

// The function that will be called to get a solution
type Terminal func(x1, x2, y1, y2 float64, img *imgut.Image)

// Functionals are internal nodes of a solution tree
type Functional func(args ...Terminal) Terminal

func (self Terminal) IsFunctional() bool {
	return false
}

func (self Terminal) Arity() int {
	return -1 // Not used
}

func (self Terminal) Run(p ...gp.Primitive) gp.Primitive {
	return self
}

func (self Functional) IsFunctional() bool {
	return true
}

func (self Functional) Arity() int {
	return 2
}

func (self Functional) Run(p ...gp.Primitive) gp.Primitive {
	return self(p[0].(Terminal), p[1].(Terminal))
}

// Buils a Terminal that fills the entire area with given color
func Filler(col ...float64) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, col...)
	}
}

// Builds a Terminal that fill the entire area with a shaded line
func LinShade(startCol, endCol, sx, sy, ex, ey float64) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.LinearShade(x1, y1, x2, y2, sx, sy, ex, ey, startCol, endCol)
	}
}

// Returns a terminal that fills the rectangle according to left and right
func VSplit(args ...Terminal) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		xh := (x1 + x2) * 0.5
		args[0](x1, y1, xh, y2, img)
		args[1](xh, y1, x2, y2, img)
	}
}

func HSplit(args ...Terminal) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		yh := (y1 + y2) * 0.5
		args[0](x1, y1, x2, yh, img)
		args[1](x1, yh, x2, y2, img)
	}
}
