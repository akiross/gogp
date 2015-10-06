package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/gpapp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/split/ts"
	"math"
)

/***********************************
 * Genetic Operators
 **********************************/

// type Terminal func(x1, x2, y float64, img *imgut.Image)

// Build terminals with names
func Black(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(0, 0, 0, 1)(x1, x2, y, img)
}

func White(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(1, 1, 1, 1)(x1, x2, y, img)
}

// Define primitives
var functionals []gogp.Primitive = []gogp.Primitive{ts.Functional(ts.Split)}
var terminals []gogp.Primitive = []gogp.Primitive{ts.Terminal(Black), ts.Terminal(White)}

func draw(ind *gpapp.Individual, img *imgut.Image) {
	// We have to compile the nodes
	exec := gogp.CompileTree(ind.Node).(ts.Terminal)
	// Apply the function
	w := float64(img.W)
	exec(1.5*w, -0.5*w, 0, img)
}

func maxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each triangle splits in 4 parts the image
	// Hence 4^n = 1, 4, 16, 64, 256... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 4^n > P -> n > log_2(P)/2
	md := int(math.Log2(float64(img.W*img.H))/2) + 1
	// Limit height to 4, to avoid trees with too many nodes
	if md > 4 {
		return 4
	}
	return md
}

func main() {
	gpapp.Evolve(maxDepth, functionals, terminals, draw)
}
