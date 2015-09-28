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
	return int(math.Log2(float64(img.W*img.H))/2) + 1
}

func main() {
	gpapp.Evolve(maxDepth, functionals, terminals, draw)
}
