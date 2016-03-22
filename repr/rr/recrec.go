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
	ephemeral  func() *Primitive
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

// Functionals returns terminals, terminals do nothing
func (p *Primitive) Run(args ...gp.Primitive) gp.Primitive {
	if p.name == "VSplit" {
		return &Primitive{"VSplit", false, -1, func(x1, y1, x2, y2 float64, img *imgut.Image) {
			xh := (x1 + x2) * 0.5
			left := args[0].(*Primitive)
			right := args[1].(*Primitive)
			left.Render(x1, y1, xh, y2, img)
			right.Render(xh, y1, x2, y2, img)
		}, nil}
	} else if p.name == "HSplit" {
		return &Primitive{"HSplit", false, -1, func(x1, y1, x2, y2 float64, img *imgut.Image) {
			yh := (y1 + y2) * 0.5
			args[0].(*Primitive).Render(x1, y1, x2, yh, img)
			args[1].(*Primitive).Render(x1, yh, x2, y2, img)
		}, nil}
	} else if p.IsEphemeral() {
		return p.ephemeral() // Generate and return new constant
	} else {
		return p
	}
}

func (p *Primitive) Name() string {
	return p.name
}

func MakeVSplit() *Primitive {
	return &Primitive{"VSplit", true, 2, nil, nil}
}

func MakeHSplit() *Primitive {
	return &Primitive{"HSplit", true, 2, nil, nil}
}

func MakeTerminal(name string, rf RenderFunc) *Primitive {
	return &Primitive{name, false, -1, rf, nil}
}

func MakeEphimeral(name string, mk func() *Primitive) *Primitive {
	return &Primitive{name, false, -1, nil, mk}
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

func DiagShade(startCol, endCol float64, diagonal bool) RenderFunc {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, startCol, startCol, startCol)
		if diagonal {
			img.DrawPoly(x1, y1, x2, y1, x2, y2)
		} else {
			img.DrawPoly(x1, y2, x2, y2, x2, y1)
		}
		img.FillColor(endCol, endCol, endCol)
	}
}

/* Make a render function for a diagonal line.
This is not a real "line", it is more a polygon, which may be very skewed.

  A +----x----+ B
    |         |
	x         x
	|         |
  C +----x----+ D

Given a diagonal, for example AD, the algorithm will draw a polygon going
from AxB to BxD and from AxC to CxD, where the point x is determined by
size. The size is a number going from 0 to 15.
Given a side, for example AB, call H half of the length of this size.
The point x on this side is determined by (size/15)*H if A is the corner where
the diagonal starts, or (1 - size/15)*H if B is the corner where the diagonal
starts.
*/
func DiagLine(bgCol, fgCol float64, diagonal bool, size int) RenderFunc {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, bgCol, bgCol, bgCol)
		fs := float64(size) / 15.0
		sDx, sDy := fs*(x2-x1)*0.5, fs*(y2-y1)*0.5
		if diagonal {
			// Diagonal from (x1,y1) to (x2,y2)
			img.DrawPoly(
				x1, y1,
				x1+sDx, y1,
				x2, y2-sDy,
				x2, y2,
				x2-sDx, y2,
				x1, y1+sDy,
			)
		} else {
			// Diagonal from (x1,y2) to (x2,y1)
			img.DrawPoly(
				x1, y2,
				x1, y2-sDy,
				x2-sDx, y1,
				x2, y1,
				x2, y1+sDy,
				x1+sDx, y2,
			)
		}
		img.FillColor(fgCol, fgCol, fgCol)
	}
}

/*
func CircShade(startCol, endCol, sx, sy, ex, ey, cx, cy, inRad, outRad float64) RenderFunc {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.CircularShade(x1, y1, x2, y2, cx, cy, inRad, outRad, startCol, endCol)
	}
}
*/
