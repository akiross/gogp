package ts

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/image/draw2d/imgut"
)

type Terminal func(x1, x2, y float64, img *imgut.Image)

type Functional func(top, left, center, right Terminal) Terminal

func (self Terminal) IsFunctional() bool {
	return false
}

func (self Terminal) Arity() int {
	return -1 // Not used
}

func (self Terminal) Run(p ...gogp.Primitive) gogp.Primitive {
	return self
}

func (self Functional) IsFunctional() bool {
	return true
}

func (self Functional) Arity() int {
	return 4
}

func (self Functional) Run(p ...gogp.Primitive) gogp.Primitive {
	return self(p[0].(Terminal), p[1].(Terminal), p[2].(Terminal), p[3].(Terminal))
}

// Return a terminal that fills the entire triangle with given color
func Filler(col ...float64) Terminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		img.FillTriangle(x1, x2, y, col...)
	}
}

func Split(top, left, center, right Terminal) Terminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		// Split the triangle in 4 parts
		cx1, cxm, cx2 := x1+0.25*(x2-x1), x1+0.5*(x2-x1), x1+0.75*(x2-x1)
		ty := imgut.TriangleCenterY(x1, x2, y)
		cy := 0.5 * (ty + y)

		top(cx1, cx2, cy, img)
		left(x1, cxm, y, img)
		center(cx2, cx1, cy, img)
		right(cxm, x2, y, img)
	}
}
